import { Box, Button, IconButton, styled } from "@mui/material";

export const HeaderBox = styled(Box)(({}) => ({
    height: "6vh",
    marginLeft: "20vw",
    marginRight: "20vw",
    display: "flex",
    alignItems: "center",
}));

export const CartButton = styled(IconButton)(({}) => ({
    marginLeft: "auto",
    marginRight: "1vw",
    color: "white",
}));

export const SlotPropsCart = {
    paper: {
        elevation: 0,
        sx: {
            overflow: "visible",
            filter: "drop-shadow(0px 2px 8px rgba(0,0,0,0.32))",
            mt: 1.5,
            "& .MuiAvatar-root": {
                width: 32,
                height: 32,
                ml: -0.5,
                mr: 1,
            },
            "&::before": {
                content: '""',
                display: "block",
                position: "absolute",
                top: 0,
                right: 14,
                width: 10,
                height: 10,
                bgcolor: "background.paper",
                transform: "translateY(-50%) rotate(45deg)",
                zIndex: 0,
            },
        },
    },
};

export const CartMenuItem = styled(Box)(({}) => ({
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    marginLeft: "1vw",
    marginRight: "1vw",
    width: "20vw",
}));

export const CartMenuTotal = styled(Box)(({}) => ({
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    marginLeft: "1vw",
    marginRight: "1vw",
    width: "20vw",
}));

export const CheckoutBox = styled(Box)(({}) => ({
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    marginLeft: "1vw",
    marginRight: "1vw",
    width: "20vw",
}));

export const SlotPropsUser = {
    paper: {
        elevation: 0,
        sx: {
            overflow: "visible",
            filter: "drop-shadow(0px 2px 8px rgba(0,0,0,0.32))",
            mt: 1.5,
            "& .MuiAvatar-root": {
                width: 32,
                height: 32,
                ml: -0.5,
                mr: 1,
            },
            "&::before": {
                content: '""',
                display: "block",
                position: "absolute",
                top: 0,
                right: 14,
                width: 10,
                height: 10,
                bgcolor: "background.paper",
                transform: "translateY(-50%) rotate(45deg)",
                zIndex: 0,
            },
        },
    },
};

export const CheckoutButton = styled(Button)(({ theme }) => ({
    alignItems: "center",
    backgroundColor: theme.palette.primary.main,
    color: "white",
    width: "20vw",
}));
