export async function createCart(): Promise<number> {
        const clientId = "Octo";
        const password = "Kerbs";

        const response = await fetch("http://localhost:8080/createCart", {
                method: "POST",
                headers: {
                        "Content-Type": "application/json",
                },
                body: JSON.stringify({
                        clientId: clientId,
                        password: password,
                }),
        });

        if (!response.ok) {
                throw new Error("Failed to create cart");
        }

        const data = await response.json();
        return data.items;
}

export async function addToCart(
        cartID: number,
        isbn: string,
        quantity: number
) {
        const response = await fetch("http://localhost:8080/addToCart", {
                method: "POST",
                headers: {
                        "Content-Type": "application/json",
                },
                body: JSON.stringify({ cartID, isbn, quantity }),
        });

        if (!response.ok) {
                throw new Error("Failed to add item to cart");
        }
        const data = await response.json();
        return data.items;
}

export async function listCart(
        cartID: number
): Promise<Record<string, number>> {
        const response = await fetch("http://localhost:8080/listCart", {
                method: "GET",
                headers: {
                        "Content-Type": "application/json",
                },
                body: JSON.stringify({ cartID }),
        });

        if (!response.ok) {
                throw new Error("Failed to fetch cart");
        }
        const data = await response.json();
        return data.items;
}

export async function checkOutCart(
        cartID: number,
        ccNumber: string,
        ccExpirationDate: Date
): Promise<Record<string, number>> {
        const response = await fetch("http://localhost:8080/checkOutCart", {
                method: "GET",
                headers: {
                        "Content-Type": "application/json",
                },
                body: JSON.stringify({ cartID, ccNumber, ccExpirationDate }),
        });
        if (!response.ok) {
                throw new Error("Failed to checkout cart");
        }
        const data = await response.json();
        return data.items;
}
