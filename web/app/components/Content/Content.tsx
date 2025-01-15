"use client";

import { useState, useEffect } from "react";
import { ContentContainer } from "./styles";
import useSnackbar from "./useSnackbar";
import Header from "../Header/Header";
import BookGrid from "../BookGrid/BookGrid";
import useCart from "./useCart";
import useCheckout from "./useCheckout";
import useCatalog from "./useCatalog";
import useUser from "./useUser";
import { Compras } from "../Compras";
import { CheckoutPopup } from "../CheckoutPopup";

export default function Content() {
    const { cart, cartID, requestCartID, requestCartItems } = useCart();
    const { snackbarState, openSnackbar, closeSnackbar } = useSnackbar(
        "top",
        "right"
    );
    const { transactionID, handleCheckout } = useCheckout(
        cartID,
        cart,
        requestCartID,
        openSnackbar
    );
    const { catalog, requestCatalog } = useCatalog();
    const { userState, purchases, requestUserPurchases, updateUserState } =
        useUser();
    const [isComprasOpen, setIsComprasOpen] = useState(false);

    useEffect(() => {
        const initialize = async () => {
            try {
                requestCatalog();
                requestCartID();
            } catch (error) {
                console.error("Initialization failed:", error);
            }
        };

        initialize();
    }, [requestCartID, requestCatalog]);

    return (
        <ContentContainer>
            <Header
                cart={cart}
                catalog={catalog}
                onOpenCompras={() => {
                    setIsComprasOpen(true);
                    requestUserPurchases();
                }}
                userState={userState}
                onUserStateChange={updateUserState}
                onCheckout={handleCheckout}
                cartID={cartID}
            />
            <Compras
                purchases={purchases}
                open={isComprasOpen}
                onClose={() => setIsComprasOpen(false)}
                catalog={catalog}
            />
            <BookGrid
                catalog={catalog}
                cartID={cartID}
                onAddToCart={requestCartItems}
            />
            <CheckoutPopup
                userState={userState}
                transactionID={transactionID}
                onClose={closeSnackbar}
                open={snackbarState.open}
                vertical={snackbarState.vertical}
                horizontal={snackbarState.horizontal}
            />
        </ContentContainer>
    );
}