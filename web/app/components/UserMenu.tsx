import History from "@mui/icons-material/History";
import Divider from "@mui/material/Divider";
import ListItemIcon from "@mui/material/ListItemIcon";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import { useRouter } from "next/navigation";
import {
    AccountCircle,
    CreditCardOff,
    EventBusy,
    NoAccounts,
} from "@mui/icons-material";

export default function UserMenu({
    anchorEl,
    open,
    handleClose,
}: {
    anchorEl: HTMLElement | null;
    open: boolean;
    handleClose: () => void;
}) {
    const router = useRouter();

    return (
        <Menu
            anchorEl={anchorEl}
            id="account-menu"
            open={open}
            onClose={handleClose}
            onClick={handleClose}
            slotProps={{
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
            }}
            transformOrigin={{ horizontal: "right", vertical: "top" }}
            anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
        >
            <MenuItem onClick={() => router.push("/miscompras")}>
                <ListItemIcon>
                    <History fontSize="small" />
                </ListItemIcon>
                Mis compras
            </MenuItem>
            <Divider />

            <MenuItem>
                <ListItemIcon>
                    <AccountCircle fontSize="small" />
                </ListItemIcon>
                Usuario Valido
            </MenuItem>
            <MenuItem>
                <ListItemIcon>
                    <NoAccounts fontSize="small" />
                </ListItemIcon>
                Usuario Invalido
            </MenuItem>
            <MenuItem>
                <ListItemIcon>
                    <EventBusy fontSize="small" />
                </ListItemIcon>
                Usuario con tarjeta expirada
            </MenuItem>
            <MenuItem>
                <ListItemIcon>
                    <CreditCardOff fontSize="small" />
                </ListItemIcon>
                Usuario sin fondos
            </MenuItem>
        </Menu>
    );
}
