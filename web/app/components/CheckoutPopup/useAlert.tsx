import Alert from "@mui/material/Alert";
import { useCallback, useState } from "react";

export const useAlert = (onClose: () => void) => {
    const [alertState, setAlertState] = useState(
        <Alert
            onClose={onClose}
            severity="warning"
            variant="filled"
            sx={{
                width: "92%",
                marginTop: "5.5vh",
                bgcolor: "#567568",
            }}
        >
            No transaction could be done!
        </Alert>
    );

    const handleState = useCallback(
        (severityCode: "error" | "success", message: string) => {
            let color = "#567568";
            if (severityCode == "error") {
                color = "#C14431";
            }
            setAlertState(
                <Alert
                    onClose={onClose}
                    severity={severityCode}
                    variant="filled"
                    sx={{
                        width: "17vw",
                        marginTop: "5.5vh",
                        bgcolor: color,
                    }}
                >
                    {message}
                </Alert>
            );
        },
        [onClose]
    );

    return { alertState, handleState };
};