package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KerbsOD/TusLibros/internal"
	"github.com/KerbsOD/TusLibros/internal/cart"
	"github.com/KerbsOD/TusLibros/internal/cashier"
	"github.com/KerbsOD/TusLibros/internal/clock"
	"github.com/KerbsOD/TusLibros/internal/merchantProcessor"
	"github.com/KerbsOD/TusLibros/internal/userAuthentication"
	"github.com/KerbsOD/TusLibros/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

func (s *HandlersTestSuite) CanNotCreateCartWithEmptyName() {
	// Given: an invalid user
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)

	// When: the cart is created without a name
	createCartResponseRecorder, createCartResponse := s.createCartRequestSender("", "Kerbs")

	// Then: Should return the invalid user error message
	assert.Equal(s.T(), http.StatusBadRequest, createCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, createCartResponse.Status)
	assert.Empty(s.T(), createCartResponse.CartID)
	assert.Equal(s.T(), internal.InvalidUserOrPasswordErrorMessage, createCartResponse.Message)
}

func (s *HandlersTestSuite) CanNotCreateCartWithEmptyPassword() {
	// Given: an invalid user
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)

	// When: the cart is created without a last name
	createCartResponseRecorder, createCartResponse := s.createCartRequestSender("Octo", "")

	// Then: throw invalid user error message
	assert.Equal(s.T(), http.StatusBadRequest, createCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, createCartResponse.Status)
	assert.Empty(s.T(), createCartResponse.CartID)
	assert.Equal(s.T(), internal.InvalidUserOrPasswordErrorMessage, createCartResponse.Message)
}

func (s *HandlersTestSuite) CanCreateCartWithValidUser() {
	// Given: A valid user and password

	// When: Creating a cart
	createCartResponseRecorder, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")

	// Then: Return valid cart
	assert.Equal(s.T(), http.StatusOK, createCartResponseRecorder.Code)
	assert.Equal(s.T(), 0, createCartResponse.Status)
	assert.Equal(s.T(), 1, createCartResponse.CartID)
	assert.Empty(s.T(), createCartResponse.Message)
}

func (s *HandlersTestSuite) DifferentIDSWhenTwoClientsRequestANewCart() {
	// Given: Two valid users

	// When: Creating a cart
	_, createCartResponse1 := s.createCartRequestSender("Octo", "Kerbs")
	_, createCartResponse2 := s.createCartRequestSender("Luca", "Zarecki")

	// Then: Cart ids are different
	assert.Equal(s.T(), 1, createCartResponse1.CartID)
	assert.Equal(s.T(), 2, createCartResponse2.CartID)
}

func (s *HandlersTestSuite) CantAddToCartWithInvalidCardID() {
	// Given: An invalid cart id

	// When: Adding to a cartid
	addToCartResponseRecorder, addToCartResponse := s.addToCartRequestSender(-1, "978-0553579901", 10)

	// Then: Can not add to invalid cart
	assert.Equal(s.T(), http.StatusBadRequest, addToCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, addToCartResponse.Status)
	assert.Equal(s.T(), internal.InvalidCartIDErrorMessage, addToCartResponse.Message)
}

func (s *HandlersTestSuite) CantAddToCartWithInvalidItem() {
	// Given: Cart with valid id
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")

	// When: Adding an invalid item
	addToCartResponseRecorder, addToCartResponse := s.addToCartRequestSender(createCartResponse.CartID, "0000", 10)

	// Then: Can not add invalid item to valid cart
	assert.Equal(s.T(), http.StatusBadRequest, addToCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, addToCartResponse.Status)
	assert.Equal(s.T(), cart.InvalidItemErrorMessage, addToCartResponse.Message)
}

func (s *HandlersTestSuite) CantAddToCartWithInvalidQuantity() {
	// Given: Cart with valid id
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")

	// When: Adding a valid item with an invalid quantity
	addToCartResponseRecorder, addToCartResponse := s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", -1)

	// Then: Can not add invalid quantity to valid cart
	assert.Equal(s.T(), http.StatusBadRequest, addToCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, addToCartResponse.Status)
	assert.Equal(s.T(), cart.InvalidQuantityErrorMessage, addToCartResponse.Message)
}

func (s *HandlersTestSuite) CanNotListInvalidCartID() {
	// Given: Invalid cart id

	// When: Listing an invalid cart
	listCartResponseRecorder, listCartResponse := s.listCartRequestSender(-1)

	// Then: Can not list an invalid cart
	assert.Equal(s.T(), http.StatusBadRequest, listCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, listCartResponse.Status)
	assert.Empty(s.T(), listCartResponse.Items)
	assert.Equal(s.T(), internal.InvalidCartIDErrorMessage, listCartResponse.Message)
}

func (s *HandlersTestSuite) CanNotListAnEmptyCart() {
	// Given: valid and empty cart
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")

	// When: Listing an empty cart
	listCartResponseRecorder, listCartResponse := s.listCartRequestSender(createCartResponse.CartID)

	// Then: Should be empty
	assert.Equal(s.T(), http.StatusBadRequest, listCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, listCartResponse.Status)
	assert.Empty(s.T(), listCartResponse.Items)
	assert.Equal(s.T(), cart.InvalidCartErrorMessage, listCartResponse.Message)
}

func (s *HandlersTestSuite) CartIsListedCorrectly() {
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

func (s *HandlersTestSuite) CantCheckOutCartWithInvalidID() {
	// Given: Invalid cart

	// When: Checking out an invalid cart
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(-1, "1111222233334444", s.tomorrow(), "Octo")

	// Then: Can not checkout invalid cart
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), internal.InvalidCartIDErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) CantCheckOutAnEmptyCart() {
	// Given: Valid cart but empty
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")

	// When: Checking out an empty cart
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.tomorrow(), "Octo")

	// Then: Cant checkout an empty cart
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), cashier.InvalidCartErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) CantCheckOutWithExpiredCreditCard() {
	// Given: valid cart and item but invalid credit card
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", 5)

	// When: Checking out with invalid credit card
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.yesterday(), "Octo")

	// Then: Cant checkout with invalid credit card
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), merchantProcessor.InvalidCreditCardErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) CantCheckOutWithInsufficientFundsCreditCard() {
	// Given: valid cart, valid item and not expired credit card
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", 5)
	s.mockMerchantProcessor.On("DebitOn", mock.Anything, mock.Anything).Unset()
	s.mockMerchantProcessor.On("DebitOn", mock.Anything, mock.Anything).Return(errors.New(merchantProcessor.InvalidCreditCardErrorMessage))

	// When: Checking out with insufficient funds
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.tomorrow(), "Octo")

	// Then: Cant checkout with no funds
	assert.Equal(s.T(), http.StatusBadRequest, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 1, checkOutCartResponse.Status)
	assert.Empty(s.T(), checkOutCartResponse.TransactionID)
	assert.Equal(s.T(), merchantProcessor.InvalidCreditCardErrorMessage, checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) CanCheckOutCorrectly() {
	// Given: valid cart and valid items
	_, createCartResponse := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse.CartID, "978-0553579901", 5)

	// When: Checking out with valid credit card
	checkOutCartResponseRecorder, checkOutCartResponse := s.checkOutCartRequestSender(createCartResponse.CartID, "1111222233334444", s.tomorrow(), "Octo")

	// Then: Can checkout
	assert.Equal(s.T(), http.StatusOK, checkOutCartResponseRecorder.Code)
	assert.Equal(s.T(), 0, checkOutCartResponse.Status)
	assert.Equal(s.T(), 1, checkOutCartResponse.TransactionID)
	assert.Empty(s.T(), checkOutCartResponse.Message)
}

func (s *HandlersTestSuite) CanNotListPurchasesOfInvalidClient() {
	// Given: an invalid client
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)

	// When: listing the purchases of an invalid client
	listPurchasesResponseRecorder, listPurchasesResponse := s.listPurchasesRequestSender("Marty", "McFly")

	// Then: Cant list purchases of an invalid client
	assert.Equal(s.T(), http.StatusBadRequest, listPurchasesResponseRecorder.Code)
	assert.Equal(s.T(), 1, listPurchasesResponse.Status)
	assert.Empty(s.T(), listPurchasesResponse.Items)
	assert.Equal(s.T(), internal.InvalidUserOrPasswordErrorMessage, listPurchasesResponse.Message)
}

func (s *HandlersTestSuite) PurchasesAreListedCorrectly() {
	// Given: Two valid carts who were checked out
	_, createCartResponse1 := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse1.CartID, "978-0553579901", 2)
	_, _ = s.checkOutCartRequestSender(createCartResponse1.CartID, "1111222233334444", s.tomorrow(), "Octo")

	_, createCartResponse2 := s.createCartRequestSender("Octo", "Kerbs")
	_, _ = s.addToCartRequestSender(createCartResponse2.CartID, "979-8712157877", 5)
	_, _ = s.checkOutCartRequestSender(createCartResponse2.CartID, "1111222233334444", s.tomorrow(), "Octo")

	// When: Listing the purchases of the client
	listPurchasesResponseRecorder, listPurchasesResponse := s.listPurchasesRequestSender("Octo", "Kerbs")

	// Then: The two purchases are present in the client account
	assert.Equal(s.T(), http.StatusOK, listPurchasesResponseRecorder.Code)
	assert.Equal(s.T(), 0, listPurchasesResponse.Status)
	assert.Equal(s.T(), map[string]float64{"978-0553579901": 39.98, "979-8712157877": 49.95}, listPurchasesResponse.Items)
	assert.Empty(s.T(), listPurchasesResponse.Message)
}

// Private

func (s *HandlersTestSuite) createCartRequestSender(clientID, password string) (*httptest.ResponseRecorder, models.CreateCartResponse) {
	body, _ := json.Marshal(models.CreateCartRequest{
		ClientID: clientID,
		Password: password,
	})
	createCartRequest := httptest.NewRequest(http.MethodPost, "/CreateCart", bytes.NewReader(body))
	createCartResponseRecorder := httptest.NewRecorder()

	s.handler.CreateCart(createCartResponseRecorder, createCartRequest)

	var createCartResponse models.CreateCartResponse
	_ = json.Unmarshal(createCartResponseRecorder.Body.Bytes(), &createCartResponse)

	return createCartResponseRecorder, createCartResponse
}

func (s *HandlersTestSuite) addToCartRequestSender(cartID int, bookISBN string, bookQuantity int) (*httptest.ResponseRecorder, models.AddToCartResponse) {
	addToCartRequestBody, _ := json.Marshal(models.AddToCartRequest{
		CartID:       cartID,
		BookISBN:     bookISBN,
		BookQuantity: bookQuantity,
	})

	addToCartRequest := httptest.NewRequest(http.MethodPost, "/AddToCart", bytes.NewReader(addToCartRequestBody))
	addToCartResponseRecorder := httptest.NewRecorder()

	s.handler.AddToCart(addToCartResponseRecorder, addToCartRequest)

	var addToCartResponse models.AddToCartResponse
	_ = json.Unmarshal(addToCartResponseRecorder.Body.Bytes(), &addToCartResponse)

	return addToCartResponseRecorder, addToCartResponse
}

func (s *HandlersTestSuite) listCartRequestSender(cartID int) (*httptest.ResponseRecorder, models.ListCartResponse) {
	listCartRequestBody, _ := json.Marshal(models.ListCartRequest{
		CartID: cartID,
	})

	listCartRequest := httptest.NewRequest(http.MethodGet, "/ListToCart", bytes.NewReader(listCartRequestBody))
	listCartResponseRecorder := httptest.NewRecorder()

	s.handler.ListCart(listCartResponseRecorder, listCartRequest)

	var listCartResponse models.ListCartResponse
	_ = json.Unmarshal(listCartResponseRecorder.Body.Bytes(), &listCartResponse)

	return listCartResponseRecorder, listCartResponse
}

func (s *HandlersTestSuite) checkOutCartRequestSender(cartID int, ccNumber string, ccExpirationDate time.Time, ccOwner string) (*httptest.ResponseRecorder, models.CheckOutCartResponse) {
	checkOutCartRequestBody, _ := json.Marshal(models.CheckOutCartRequest{
		CartID:                   cartID,
		CreditCardNumber:         ccNumber,
		CreditCardExpirationDate: ccExpirationDate,
		CreditCardOwner:          ccOwner,
	})

	checkOutCartRequest := httptest.NewRequest(http.MethodPost, "/CheckOutCart", bytes.NewReader(checkOutCartRequestBody))
	checkOutCartResponseRecorder := httptest.NewRecorder()

	s.handler.CheckOutCart(checkOutCartResponseRecorder, checkOutCartRequest)

	var checkOutCartResponse models.CheckOutCartResponse
	_ = json.Unmarshal(checkOutCartResponseRecorder.Body.Bytes(), &checkOutCartResponse)

	return checkOutCartResponseRecorder, checkOutCartResponse
}

func (s *HandlersTestSuite) listPurchasesRequestSender(clientID, password string) (*httptest.ResponseRecorder, models.ListPurchasesResponse) {
	body, _ := json.Marshal(models.ListPurchasesRequest{
		ClientID: clientID,
		Password: password,
	})
	listPurchasesRequest := httptest.NewRequest(http.MethodGet, "/ListPurchases", bytes.NewReader(body))
	listPurchasesResponseRecorder := httptest.NewRecorder()

	s.handler.ListPurchases(listPurchasesResponseRecorder, listPurchasesRequest)

	var listPurchasesResponse models.ListPurchasesResponse
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
