import { Alert, Snackbar } from "@mui/material";
import React, { useEffect, useState } from "react";
import { UserState } from "./types";

export default function CheckoutPopup({
    userState,
    onClose,
    open,
    vertical,
    horizontal,
}: {
    userState: UserState;
    onClose: () => void;
    open: boolean;
    vertical: "bottom" | "top";
    horizontal: "center" | "left" | "right";
}) {
    const [alertState, setAlertState] = useState(
        <Alert
            onClose={onClose}
            severity="success"
            variant="filled"
            sx={{
                width: "92%",
                marginTop: "5.5vh",
                bgcolor: "#567568",
            }}
        >
            Transaction completed succesfully!
        </Alert>
    );

    useEffect(() => {
        switch (userState) {
            case UserState.InvalidUser:
                setAlertState(
                    <Alert
                        onClose={onClose}
                        severity="error"
                        variant="filled"
                        sx={{
                            width: "17vw",
                            marginTop: "5.5vh",
                        }}
                    >
                        Can not check out, invalid user!
                    </Alert>
                );
                break;

            case UserState.ExpiredCreditCardUser:
                setAlertState(
                    <Alert
                        onClose={onClose}
                        severity="error"
                        variant="filled"
                        sx={{
                            width: "17vw",
                            marginTop: "5.5vh",
                        }}
                    >
                        Can not check out, the used credit card is expired!
                    </Alert>
                );
                break;

            case UserState.NoFundsCreditCardUser:
                setAlertState(
                    <Alert
                        onClose={onClose}
                        severity="error"
                        variant="filled"
                        sx={{
                            width: "17vw",
                            marginTop: "5.5vh",
                        }}
                    >
                        Can not check out, the credit card has insufficient
                        funds!
                    </Alert>
                );
                break;

            default:
                setAlertState(
                    <Alert
                        onClose={onClose}
                        severity="success"
                        variant="filled"
                        sx={{
                            width: "17vw",
                            marginTop: "5.5vh",
                            bgcolor: "#567568",
                        }}
                    >
                        Transaction completed succesfully!
                    </Alert>
                );
                break;
        }
    }, [onClose, userState]);

    return (
        <div>
            <Snackbar
                anchorOrigin={{ vertical, horizontal }}
                open={open}
                autoHideDuration={6000}
                onClose={onClose}
                key={vertical + horizontal}
            >
                {alertState}
            </Snackbar>
        </div>
    );
}
