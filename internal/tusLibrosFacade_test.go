package internal

import (
	"errors"
	"testing"
	"time"

	"github.com/KerbsOD/TusLibros/internal/cart"
	"github.com/KerbsOD/TusLibros/internal/cashier"
	"github.com/KerbsOD/TusLibros/internal/clock"
	"github.com/KerbsOD/TusLibros/internal/merchantProcessor"
	"github.com/KerbsOD/TusLibros/internal/testsObjectFactory"
	"github.com/KerbsOD/TusLibros/internal/userAuthentication"
	"github.com/KerbsOD/TusLibros/internal/userCredentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type FacadeTestSuite struct {
	suite.Suite
	factory                testsObjectFactory.TestsObjectFactory
	catalog                map[string]float64
	book1                  string
	book1Price             float64
	book2                  string
	book2Price             float64
	invalidBook            string
	validUserCredentials   *userCredentials.UserCredentials
	invalidUserCredentials *userCredentials.UserCredentials
	yesterday              time.Time
	today                  time.Time
	tomorrow               time.Time
	validCardNumber        string
	validCardOwner         string
	mockMerchantProcessor  *merchantProcessor.MockMerchantProcessor
	mockClock              *clock.MockClock
	mockUserAuthentication *userAuthentication.MockUserAuthentication
	facade                 *SystemFacade
}

func TestFacadeTestSuite(t *testing.T) {
	suite.Run(t, new(FacadeTestSuite))
}

func (s *FacadeTestSuite) SetupTest() {
	s.factory = testsObjectFactory.TestsObjectFactory{}
	s.catalog = s.factory.CatalogWithAnItemAndItsPrice()
	s.book1 = s.factory.ItemInCatalog()
	s.book1Price = s.factory.ItemInCatalogPrice()
	s.book2 = s.factory.AnotherItemInCatalog()
	s.book2Price = s.factory.AnotherItemInCatalogPrice()
	s.invalidBook = s.factory.ItemNotInCatalog()
	s.validUserCredentials = s.factory.ValidUserCredentials()
	s.invalidUserCredentials = s.factory.InvalidUserCredentials()
	s.yesterday = s.factory.Yesterday()
	s.today = s.factory.Today()
	s.tomorrow = s.factory.Tomorrow()
	s.validCardNumber = s.factory.ValidCreditCardNumber()
	s.validCardOwner = s.factory.ValidCreditCardOwner()
	s.mockMerchantProcessor = merchantProcessor.NewMockMerchantProcessor()
	s.mockClock = clock.NewMockClock()
	s.mockUserAuthentication = userAuthentication.NewMockUserAuthentication()
	s.facade = NewSystemFacade(s.catalog, s.mockUserAuthentication, s.mockMerchantProcessor, s.mockClock)
}

func (s *FacadeTestSuite) Test01CanCreateCartWithValidUsernameAndValidUsernamePassword() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	myCart, _ := s.facade.CartWithID(cartID)
	assert.True(s.T(), myCart.IsEmpty())
}

func (s *FacadeTestSuite) Test02CanNotCreateCartWithInvalidUsername() {
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)
	_, err := s.facade.CreateCart(s.invalidUserCredentials)
	assert.EqualError(s.T(), err, InvalidUserOrPasswordErrorMessage)
}

func (s *FacadeTestSuite) Test03CanNotCreateCartWithInvalidUsernamePassword() {
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)
	_, err := s.facade.CreateCart(s.invalidUserCredentials)
	assert.EqualError(s.T(), err, InvalidUserOrPasswordErrorMessage)
}

func (s *FacadeTestSuite) Test04CanAddItemsToACreatedCart() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	myCart, _ := s.facade.CartWithID(cartID)
	assert.False(s.T(), myCart.IsEmpty())
}

func (s *FacadeTestSuite) Test05CanNotAddItemToNotCreatedCart() {
	err := s.facade.AddToCart(-1, s.book1, 10)
	assert.EqualError(s.T(), err, InvalidCartIDErrorMessage)
}

func (s *FacadeTestSuite) Test06CanNotAddItemNotSellByTheStore() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	err := s.facade.AddToCart(cartID, s.invalidBook, 10)
	assert.EqualError(s.T(), err, cart.InvalidItemErrorMessage)
}

func (s *FacadeTestSuite) Test07CanNotAddNegativeQuantityOfAnItem() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	err := s.facade.AddToCart(cartID, s.book1, -1)
	assert.EqualError(s.T(), err, cart.InvalidQuantityErrorMessage)
}

func (s *FacadeTestSuite) Test08ListCartOfAnEmptyCartReturnsEmptyMap() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	addedItems, _ := s.facade.ListCart(cartID)
	assert.ElementsMatch(s.T(), map[string]int{}, addedItems)
}

func (s *FacadeTestSuite) Test09CanNotListCartWithInvalidCartID() {
	_, err := s.facade.ListCart(-1)
	assert.EqualError(s.T(), err, InvalidCartIDErrorMessage)
}

func (s *FacadeTestSuite) Test10ListCartListsAddedItemsCorrectly() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	addedItems, _ := s.facade.ListCart(cartID)
	expectedItems := map[string]int{}
	expectedItems[s.book1] = 10
	assert.Equal(s.T(), expectedItems, addedItems)
}

func (s *FacadeTestSuite) Test11CanCheckOutACart() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	_, _ = s.facade.CheckOutCart(cartID, s.validCardNumber, s.tomorrow, s.validCardOwner)
	clientPurchases, _ := s.facade.ListPurchasesOf(s.validUserCredentials)
	expectedPurchases := map[string]float64{}
	expectedPurchases[s.book1] = s.book1Price * 10
	assert.Equal(s.T(), expectedPurchases, clientPurchases)
}

func (s *FacadeTestSuite) Test12CanNotCheckoutCartWithInvalidCartID() {
	_, err := s.facade.CheckOutCart(-1, s.validCardNumber, s.tomorrow, s.validCardOwner)
	assert.EqualError(s.T(), err, InvalidCartIDErrorMessage)
}

func (s *FacadeTestSuite) Test13CanNotCheckoutEmptyCart() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_, err := s.facade.CheckOutCart(cartID, s.validCardNumber, s.tomorrow, s.validCardOwner)
	assert.EqualError(s.T(), err, cashier.InvalidCartErrorMessage)
}

func (s *FacadeTestSuite) Test14CanNotCheckoutWithAnExpiredCreditCard() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	_, err := s.facade.CheckOutCart(cartID, s.validCardNumber, s.yesterday, s.validCardOwner)
	assert.EqualError(s.T(), err, merchantProcessor.InvalidCreditCardErrorMessage)
}

func (s *FacadeTestSuite) Test15CanNotCheckoutWithInsufficientFunds() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	s.mockMerchantProcessor.On("DebitOn", mock.Anything, mock.Anything).Unset()
	s.mockMerchantProcessor.On("DebitOn", mock.Anything, mock.Anything).Return(errors.New(merchantProcessor.InvalidCreditCardErrorMessage))
	_, err := s.facade.CheckOutCart(cartID, s.validCardNumber, s.tomorrow, s.validCardOwner)
	assert.EqualError(s.T(), err, merchantProcessor.InvalidCreditCardErrorMessage)
}

func (s *FacadeTestSuite) Test16PurchaseIsRegisteredInSalesBook() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	_ = s.facade.AddToCart(cartID, s.book2, 5)
	_, _ = s.facade.CheckOutCart(cartID, s.validCardNumber, s.tomorrow, s.validCardOwner)
	expectedPurchases := map[string]float64{}
	expectedPurchases[s.book1] = s.book1Price * 10
	expectedPurchases[s.book2] = s.book2Price * 5
	clientPurchases, _ := s.facade.ListPurchasesOf(s.validUserCredentials)
	assert.Equal(s.T(), expectedPurchases, clientPurchases)
}

func (s *FacadeTestSuite) Test17CanNotListPurchasesOfInvalidUsername() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	_, _ = s.facade.CheckOutCart(cartID, s.validCardNumber, s.tomorrow, s.validCardOwner)
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)
	_, err := s.facade.ListPurchasesOf(s.invalidUserCredentials)
	assert.EqualError(s.T(), err, InvalidUserOrPasswordErrorMessage)
}

func (s *FacadeTestSuite) Test18CanNotListPurchasesOfUsernameWithInvalidPassword() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	_, _ = s.facade.CheckOutCart(cartID, s.validCardNumber, s.tomorrow, s.validCardOwner)
	s.mockUserAuthentication.On("RegisteredUser", mock.Anything, mock.Anything).Return(false)
	_, err := s.facade.ListPurchasesOf(s.invalidUserCredentials)
	assert.EqualError(s.T(), err, InvalidUserOrPasswordErrorMessage)
}

func (s *FacadeTestSuite) Test19CanNotAddToCartWhenSessionHasExpired() {
	currentTime := s.today
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	currentTime = currentTime.Add(31 * time.Minute)
	s.mockClock.On("Now").Return(currentTime)
	err := s.facade.AddToCart(cartID, s.book1, 10)
	assert.EqualError(s.T(), err, InvalidCartIDErrorMessage)
}

func (s *FacadeTestSuite) Test20CanNotListCartWhenSessionHasExpired() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	nowPlus30Minutes := s.today.Add(31 * time.Minute)
	s.mockClock.On("Now").Return(nowPlus30Minutes)
	_, err := s.facade.ListCart(cartID)
	assert.EqualError(s.T(), err, InvalidCartIDErrorMessage)
}

func (s *FacadeTestSuite) Test21CanNotCheckoutCartWhenSessionHasExpired() {
	currentTime := s.today
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	_ = s.facade.AddToCart(cartID, s.book1, 10)
	currentTime = currentTime.Add(31 * time.Minute)
	s.mockClock.On("Now").Return(currentTime)
	_, err := s.facade.CheckOutCart(cartID, s.validCardNumber, s.tomorrow, s.validCardOwner)
	assert.EqualError(s.T(), err, InvalidCartIDErrorMessage)
}

func (s *FacadeTestSuite) Test22UsingCartUpdatesLastUsedTimeSoItDoesntExpire() {
	cartID, _ := s.facade.CreateCart(s.validUserCredentials)
	currentTime := s.today

	// Esperamos 20 minutos y agregamos un libro
	currentTime = currentTime.Add(20 * time.Minute)
	s.mockClock.On("Now").Return(currentTime).Once()
	_ = s.facade.AddToCart(cartID, s.book1, 10)

	// Esperamos otros 20 minutos y decidimos listar el carrito
	currentTime = currentTime.Add(20 * time.Minute)
	s.mockClock.On("Now").Return(currentTime).Once()
	listedItems, _ := s.facade.ListCart(cartID)

	expectedItems := map[string]int{}
	expectedItems[s.book1] = 10
	assert.Equal(s.T(), expectedItems, listedItems)
}
