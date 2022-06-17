import { useCallback, useEffect, useRef, useState } from "react";
import { ApiResponse, OrderT, PrinterT, ProductT, StationT, UserT } from './types';

type GetProductsStateT = {
    products: ProductT[];
    loading: boolean;
    socketIsOpen: boolean;
    waitingToReconnect: boolean;
};

/**
 * A hook that gets and returns the list of available products.
 * @returns The list of available products.
 */
export const useGetProducts = () => {
    const [state, setState] = useState<GetProductsStateT>({
        products: [],
        loading: false,
        socketIsOpen: false,
        waitingToReconnect: false,
    });
    const stateRef = useRef(state);
    const wsRef = useRef<WebSocket | null>(null);
    const waitRef = useRef<NodeJS.Timer | null>(null);

    const onSocketMessage = async (event: MessageEvent<any>) => {
        const data = JSON.parse(await event.data.text()) as ProductT;

        if (stateRef.current.products.length > 0) {
            const products = [...stateRef.current.products];
            const index = products.findIndex((p) => p.id === data.id);

            if (index < 0) {
                products.push(data);
            } else {
                products[index] = data;
            }

            setState({
                ...stateRef.current,
                products: products,
            });
        }
    };
    const onSocketOpen = () => {
        setState({
            ...stateRef.current,
            socketIsOpen: true,
        });
        console.log('SOCKET_CONNECTED', new Date());
    };
    const onSocketClose = () => {
        if (stateRef.current.waitingToReconnect || !wsRef.current) {
            return;
        }

        wsRef.current = null;
        setState({
            ...stateRef.current,
            socketIsOpen: false,
            waitingToReconnect: true,
        });

        if (!!waitRef.current) {
            return;
        }

        // Try to reconnect every second.
        waitRef.current = setInterval(() => {
            if (stateRef.current.socketIsOpen) {
                // @ts-ignore
                clearInterval(waitRef.current);
                return;
            }

            // Remove the waitingToReconnect flag so that the client can reconnect.
            if (stateRef.current.waitingToReconnect) {
                setState({
                    ...stateRef.current,
                    waitingToReconnect: false,
                });
            }
        }, 1000);
    };

    useEffect(() => {
        // This is what submits the actual request to get the list of products.
        const getProducts = async () => {
            const req = await fetch('/api/products');
            const resp = await req.json();

            if (resp.success) {
                setState({
                    ...stateRef.current,
                    products: resp.data,
                    loading: false,
                });
            }
        };

        getProducts();
        setState({
            ...stateRef.current,
            loading: true,
        });
    }, []);

    useEffect(() => {
        if (state.waitingToReconnect || !!wsRef.current) {
            return;
        }

        const host = window.location.host;
        const socket = new WebSocket(`ws://${host}/api/products/ws`);
        wsRef.current = socket;

        socket.onmessage = onSocketMessage;
        socket.onopen = onSocketOpen;
        socket.onerror = (e) => console.log('SOCKET_ERROR', e);
        socket.onclose = onSocketClose;

        return () => {
            wsRef.current?.close();
            wsRef.current = null;
        };
    }, [state.waitingToReconnect]);

    useEffect(() => {
        stateRef.current = state;
    }, [state]);

    return state;
};

type CreateOrderStateT = {
    order: OrderT | null;
    loading: boolean;
};

/**
 * Hook that handles the creation of an order.
 */
export const useCreateOrder = () => {
    const [ state, setState ] = useState<CreateOrderStateT>({
        order: null,
        loading: false,
    });

    /**
     * Creates the order.
     * @param productIds The ids to attach to the new order.
     * @returns The order
     */
    const createOrder = async (productIds: number[]): Promise<OrderT | null> => {
        setState({ ...state, loading: true });

        const body = new FormData();
        body.append('products', JSON.stringify(productIds));

        const req = await fetch('/api/order', {
            method: 'POST',
            body,
        });
        const resp = await req.json();

        if (!resp.success) {
            setState({ ...state, loading: false });
            return null;
        }

        setState({ order: resp.data, loading: false });

        return resp.data;
    };

    return { ...state, createOrder };
};

/**
 * This custom hook retrieves the list of printers.
 * @returns The printers and a function to manually get them.
 */
export const useGetPrinters = () => {
    const [printers, setPrinters] = useState<PrinterT[]>([]);

    const getPrinters = async () => {
        const req = await fetch('/api/printers');
        const resp = await req.json();

        setPrinters(resp.data);
    };

    useEffect(() => {
        getPrinters();
    }, []);

    return { printers, getPrinters };
};

type UserStateT = {
    user: UserT | undefined;
    loading: boolean;
    error: string | undefined;
}

/**
 * A hook that returns the currently logged in user.
 * @returns The user and the function to manually get the user.
 */
export const useGetUser = () => {
    const [state, setState] = useState<UserStateT>({
        user: undefined,
        loading: true,
        error: undefined,
    });

    const getUser = useCallback(async () => {
        const req = await fetch('/api/user');
        const resp = await req.json() as ApiResponse<UserT | null>;

        setState({
            user: resp.data || undefined,
            loading: false,
            error: resp.error,
        });

        return resp.data;
    }, []);

    useEffect(() => {
        getUser();
    }, []);

    return { ...state, getUser };
};
