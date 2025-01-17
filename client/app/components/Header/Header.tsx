"use client";

import { AppBar, IconButton, Tooltip, Badge, useTheme } from "@mui/material";
import { ShoppingCart } from "@mui/icons-material";
import React from "react";
import CartMenu from "./CartMenu";
import { Book } from "../Types/cart";
import { SnackbarState, User, UserState } from "../Types/user";
import { useHeaderLogic } from "./useHeaderLogic";
import { CartButton, HeaderBox } from "./styles";
import Title from "./Title";
import UserMenu from "./UserMenu";

export default function Header({
        cart,
        catalog,
        onOpenCompras,
        user,
        onUserStateChange,
        onCheckout,
}: {
        cart: Record<string, number>;
        catalog: Record<string, Book>;
        onOpenCompras: () => void;
        user: User;
        onUserStateChange: (newState: UserState) => void;
        onCheckout: (
                position: Pick<SnackbarState, "vertical" | "horizontal">
        ) => void;
}) {
        const theme = useTheme();
        const { anchorElCart, anchorElUser, handleClick, handleClose } =
                useHeaderLogic();

        return (
                <AppBar
                        position="fixed"
                        sx={{ bgcolor: theme.palette.primary.main }}
                >
                        <HeaderBox>
                                <Title />

                                <Tooltip title="List cart">
                                        <CartButton
                                                onClick={handleClick("cart")}
                                                aria-controls={
                                                        anchorElCart
                                                                ? "cart-menu"
                                                                : undefined
                                                }
                                                aria-haspopup="true"
                                                aria-expanded={Boolean(
                                                        anchorElCart
                                                )}
                                        >
                                                <Badge
                                                        badgeContent={
                                                                Object.keys(
                                                                        cart
                                                                ).length
                                                        }
                                                        color="error"
                                                >
                                                        <ShoppingCart />
                                                </Badge>
                                        </CartButton>
                                </Tooltip>

                                <Tooltip title="Account settings">
                                        <IconButton
                                                onClick={handleClick("user")}
                                                aria-controls={
                                                        anchorElUser
                                                                ? "user-menu"
                                                                : undefined
                                                }
                                                aria-haspopup="true"
                                                aria-expanded={Boolean(
                                                        anchorElUser
                                                )}
                                                sx={{ color: "white" }}
                                        >
                                                {user.logo}
                                        </IconButton>
                                </Tooltip>
                        </HeaderBox>

                        <CartMenu
                                anchorEl={anchorElCart}
                                catalog={catalog}
                                open={Boolean(anchorElCart)}
                                handleClose={handleClose("cart")}
                                cart={cart}
                                onCheckout={onCheckout}
                        />

                        <UserMenu
                                anchorEl={anchorElUser}
                                open={Boolean(anchorElUser)}
                                handleClose={handleClose("user")}
                                onUserStateChange={onUserStateChange}
                                onOpenCompras={onOpenCompras}
                        />
                </AppBar>
        );
}
