import Box from "@mui/material/Box";
import { Metadata } from "next";

export const metadata: Metadata = {
    title: "TusLibros - Mis Compras",
};

export default function Home() {
    return (
        <Box
            sx={{
                bgcolor: "#F3FCF0",
                width: "100vw",
                height: "100vh",
                overflow: "auto",
            }}
        ></Box>
    );
}