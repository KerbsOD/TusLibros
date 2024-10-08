package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/KerbsOD/TusLibros/internal"
	"github.com/KerbsOD/TusLibros/internal/cart"
	"github.com/KerbsOD/TusLibros/internal/cashier"
	"github.com/KerbsOD/TusLibros/internal/clock"
	"github.com/KerbsOD/TusLibros/internal/merchantProcessor"
	"github.com/KerbsOD/TusLibros/internal/userAuthentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type HandlersTestSuite struct {
	suite.Suite
	catalog                map[string]float64
	mockMerchantProcessor  *merchantProcessor.MockMerchantProcessor
	mockClock              *clock.MockClock
	mockUserAuthentication *userAuthentication.MockUserAuthentication
	facade                 *internal.SystemFacade
	handler                *Handler
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (s *HandlersTestSuite) SetupTest() {
	s.catalog = map[string]float64{"978-0553579901": 19.99, "979-8712157877": 9.99}
	s.mockMerchantProcessor = merchantProcessor.NewMockMerchantProcessor()
	s.mockMerchantProcessor.On("DebitOn", mock.Anything, mock.Anything).Return(nil)
	s.mockClock = clock.NewMockClock()
	s.mockUserAuthentication = userAuthentication.NewMockUserAuthentication()
	s.facade = internal.NewSystemFacade(s.catalog, s.mockUserAuthentication, s.mockMerchantProcessor, s.mockClock)
	s.handler = &Handler{Facade: s.facade}
}

func (s *HandlersTestSuite) Test01CanNotCreateCartWithEmptyName() {
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)
	createCartResponseRecorder, createCartResponse := s.createCartRequestSender("", "Kerbs")
	assert.Equal(s.T(), http.StatusBadRequest, createCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, createCartResponse.Status)
	assert.Empty(s.T(), createCartResponse.CartID)
	assert.Equal(s.T(), internal.InvalidUserOrPasswordErrorMessage, createCartResponse.Message)
}

func (s *HandlersTestSuite) Test02CanNotCreateCartWithEmptyPassword() {
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)
	createCartResponseRecorder, createCartResponse := s.createCartRequestSender("Octo", "")
	assert.Equal(s.T(), http.StatusBadRequest, createCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, createCartResponse.Status)
	assert.Empty(s.T(), createCartResponse.CartID)
	assert.Equal(s.T(), internal.InvalidUserOrPasswordErrorMessage, createCartResponse.Message)
}

func (s *HandlersTestSuite) Test03CanCreateCartWithValidUser() {
	createCartResponseRecorder, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	assert.Equal(s.T(), http.StatusOK, createCartResponseRecorder.Code)
	assert.Equal(s.T(), 0, createCartResponse.Status)
	assert.Equal(s.T(), 1, createCartResponse.CartID)
	assert.Empty(s.T(), createCartResponse.Message)
}

func (s *HandlersTestSuite) Test04DifferentIDSWhenTwoClientsRequestANewCart() {
	_, createCartResponse1 := s.createCartRequestSender("Octo", "Kerbs")
	_, createCartResponse2 := s.createCartRequestSender("Luca", "Zarecki")
	assert.Equal(s.T(), 1, createCartResponse1.CartID)
	assert.Equal(s.T(), 2, createCartResponse2.CartID)
}

func (s *HandlersTestSuite) Test05CantAddToCartWithInvalidCardID() {
	addToCartResponseRecorder, addToCartResponse := s.addToCartRequestSender(-1, "978-0553579901", 10)
	assert.Equal(s.T(), http.StatusBadRequest, addToCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, addToCartResponse.Status)
	assert.Equal(s.T(), internal.InvalidCartIDErrorMessage, addToCartResponse.Message)
}

func (s *HandlersTestSuite) Test06CantAddToCartWithInvalidItem() {
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	addToCartResponseRecorder, addToCartResponse := s.addToCartRequestSender(createCartResponse.CartID, "0000", 10)
	assert.Equal(s.T(), http.StatusBadRequest, addToCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, addToCartResponse.Status)
	assert.Equal(s.T(), cart.InvalidItemErrorMessage, addToCartResponse.Message)
}

func (s *HandlersTestSuite) Test07CantAddToCartWithInvalidQuantity() {
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	addToCartResponseRecorder, addToCartResponse := s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", -1)
	assert.Equal(s.T(), http.StatusBadRequest, addToCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, addToCartResponse.Status)
	assert.Equal(s.T(), cart.InvalidQuantityErrorMessage, addToCartResponse.Message)
}

func (s *HandlersTestSuite) Test08CanNotListInvalidCartID() {
	listCartResponseRecorder, listCartResponse := s.listCartRequestSender(-1)
	assert.Equal(s.T(), http.StatusBadRequest, listCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, listCartResponse.Status)
	assert.Empty(s.T(), listCartResponse.Items)
	assert.Equal(s.T(), internal.InvalidCartIDErrorMessage, listCartResponse.Message)
}

func (s *HandlersTestSuite) Test09CanNotListAnEmptyCart() {
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	listCartResponseRecorder, listCartResponse := s.listCartRequestSender(createCartResponse.CartID)
	assert.Equal(s.T(), http.StatusBadRequest, listCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, listCartResponse.Status)
	assert.Empty(s.T(), listCartResponse.Items)
	assert.Equal(s.T(), cart.InvalidCartErrorMessage, listCartResponse.Message)
}

func (s *HandlersTestSuite) Test10CartIsListedCorrectly() {
    // Given: A cart is created and items are added to it
    _, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
    _, _ = s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", 5)
    _, _ = s.addToCartRequestSender(createCartResponse.CartID, "979-8712157877", 10)

    // When: The cart is listed 
    listCartResponseRecorder, listCartResponse := s.listCartRequestSender(createCartResponse.CartID)

    // Then: The cart should be listed correctly with the correct status and items
    assert.Equal(s.T(), http.StatusOK, listCartResponseRecorder.Code)
    assert.Equal(s.T(), 0, listCartResponse.Status)
    assert.Equal(s.T(), map[string]int{"978-0553579901": 5, "979-8712157877": 10}, listCartResponse.Items)
    assert.Empty(s.T(), listCartResponse.Message)
}

func (s *HandlersTestSuite) Test11CantCheckOutCartWithInvalidID() {
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(-1, "1111222233334444", s.tomorrow(), "Octo")
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), internal.InvalidCartIDErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) Test12CantCheckOutAnEmptyCart() {
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.tomorrow(), "Octo")
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), cashier.InvalidCartErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) Test13CantCheckOutWithExpiredCreditCard() {
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", 5)
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.yesterday(), "Octo")
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), merchantProcessor.InvalidCreditCardErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) Test14CantCheckOutWithInsufficientFundsCreditCard() {
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", 5)
	s.mockMerchantProcessor.On("DebitOn", mock.Anything, mock.Anything).Unset()
	s.mockMerchantProcessor.On("DebitOn", mock.Anything, mock.Anything).Return(errors.New(merchantProcessor.InvalidCreditCardErrorMessage))
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.tomorrow(), "Octo")
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), merchantProcessor.InvalidCreditCardErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) Test15CanCheckOutCorrectly() {
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", 5)
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.tomorrow(), "Octo")
	assert.Equal(s.T(), http.StatusOK, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 0, checkOutCartResponse.Status)
	assert.Equal(s.T(), 1, checkOutCartResponse.TransactionID)
	assert.Empty(s.T(), checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) Test16CanNotListPurchasesOfInvalidClient() {
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)
	listPurchasesResponseRecorder, listPurchasesResponse := s.listPurchasesRequestSender("Marty", "McFly")
	assert.Equal(s.T(), http.StatusBadRequest, listPurchasesResponseRecorder.Code)
	assert.Equal(s.T(), 1, listPurchasesResponse.Status)
	assert.Empty(s.T(), listPurchasesResponse.Items)
	assert.Equal(s.T(), internal.InvalidUserOrPasswordErrorMessage, listPurchasesResponse.Message)
}

func (s *HandlersTestSuite) Test17PurchasesAreListedCorrectly() {
	_, createCartResponse1 := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse1.CartID, "978-0553579901", 2)
	_, _ = s.checkOutCartRequestSender(createCartResponse1.CartID, "1111222233334444", s.tomorrow(), "Octo")

	_, createCartResponse2 := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse2.CartID, "979-8712157877", 5)
	_, _ = s.checkOutCartRequestSender(createCartResponse2.CartID, "1111222233334444", s.tomorrow(), "Octo")

	listPurchasesResponseRecorder, listPurchasesResponse := s.listPurchasesRequestSender("Octo", "Kerbs")

	assert.Equal(s.T(), http.StatusOK, listPurchasesResponseRecorder.Code)
	assert.Equal(s.T(), 0, listPurchasesResponse.Status)
	assert.Equal(s.T(), map[string]float64{"978-0553579901": 39.98, "979-8712157877": 49.95}, listPurchasesResponse.Items)
	assert.Empty(s.T(), listPurchasesResponse.Message)
}

// Private

func (s *HandlersTestSuite) createCartRequestSender(clientID, password string) (*httptest.ResponseRecorder, CreateCartResponse) {
	body, _ := json.Marshal(CreateCartRequest{
		ClientID: clientID,
		Password: password,
	})
	createCartRequest := httptest.NewRequest(http.MethodPost, "/CreateCart", bytes.NewReader(body))
	createCartResponseRecorder := httptest.NewRecorder()

	s.handler.CreateCart(createCartResponseRecorder, createCartRequest)

	var createCartResponse CreateCartResponse
	_ = json.Unmarshal(createCartResponseRecorder.Body.Bytes(), &createCartResponse)

	return createCartResponseRecorder, createCartResponse
}

func (s *HandlersTestSuite) addToCartRequestSender(cartID int, bookISBN string, bookQuantity int) (*httptest.ResponseRecorder, AddToCartResponse) {
	addToCartRequestBody, _ := json.Marshal(AddToCartRequest{
		CartID:       cartID,
		BookISBN:     bookISBN,
		BookQuantity: bookQuantity,
	})

	addToCartRequest := httptest.NewRequest(http.MethodPost, "/AddToCart", bytes.NewReader(addToCartRequestBody))
	addToCartResponseRecorder := httptest.NewRecorder()

	s.handler.AddToCart(addToCartResponseRecorder, addToCartRequest)

	var addToCartResponse AddToCartResponse
	_ = json.Unmarshal(addToCartResponseRecorder.Body.Bytes(), &addToCartResponse)

	return addToCartResponseRecorder, addToCartResponse
}

func (s *HandlersTestSuite) listCartRequestSender(cartID int) (*httptest.ResponseRecorder, ListCartResponse) {
	listCartRequestBody, _ := json.Marshal(ListCartRequest{
		CartID: cartID,
	})

	listCartRequest := httptest.NewRequest(http.MethodGet, "/ListToCart", bytes.NewReader(listCartRequestBody))
	listCartResponseRecorder := httptest.NewRecorder()

	s.handler.ListCart(listCartResponseRecorder, listCartRequest)

	var listCartResponse ListCartResponse
	_ = json.Unmarshal(listCartResponseRecorder.Body.Bytes(), &listCartResponse)

	return listCartResponseRecorder, listCartResponse
}

func (s *HandlersTestSuite) checkOutCartRequestSender(cartID int, ccNumber string, ccExpirationDate time.Time, ccOwner string) (*httptest.ResponseRecorder, CheckOutCartResponse) {
	checkOutCartRequestBody, _ := json.Marshal(CheckOutCartRequest{
		CartID:                   cartID,
		CreditCardNumber:         ccNumber,
		CreditCardExpirationDate: ccExpirationDate,
		CreditCardOwner:          ccOwner,
	})

	checkOutCartRequest := httptest.NewRequest(http.MethodPost, "/CheckOutCart", bytes.NewReader(checkOutCartRequestBody))
	checkOutCartResponseRecorder := httptest.NewRecorder()

	s.handler.CheckOutCart(checkOutCartResponseRecorder, checkOutCartRequest)

	var checkOutCartResponse CheckOutCartResponse
	_ = json.Unmarshal(checkOutCartResponseRecorder.Body.Bytes(), &checkOutCartResponse)

	return checkOutCartResponseRecorder, checkOutCartResponse
}

func (s *HandlersTestSuite) listPurchasesRequestSender(clientID, password string) (*httptest.ResponseRecorder, ListPurchasesResponse) {
	body, _ := json.Marshal(ListPurchasesRequest{
		ClientID: clientID,
		Password: password,
	})
	listPurchasesRequest := httptest.NewRequest(http.MethodGet, "/ListPurchases", bytes.NewReader(body))
	listPurchasesResponseRecorder := httptest.NewRecorder()

	s.handler.ListPurchases(listPurchasesResponseRecorder, listPurchasesRequest)

	var listPurchasesResponse ListPurchasesResponse
	_ = json.Unmarshal(listPurchasesResponseRecorder.Body.Bytes(), &listPurchasesResponse)

	return listPurchasesResponseRecorder, listPurchasesResponse
}

func (s *HandlersTestSuite) tomorrow() time.Time {
	dayAfterStaticDateForTesting := s.mockClock.Now().AddDate(0, 0, 1)
	return dayAfterStaticDateForTesting
}

func (s *HandlersTestSuite) yesterday() time.Time {
	dayAfterStaticDateForTesting := s.mockClock.Now().AddDate(0, 0, -1)
	return dayAfterStaticDateForTesting
}
