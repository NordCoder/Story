import axios, { AxiosHeaders } from 'axios';
import { useAuth } from '@/hooks/useAuth';

const api = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL,
    withCredentials: true,
});

api.interceptors.request.use(config => {
    const token = useAuth.getState().accessToken;
    if (token) {
        const headers = config.headers instanceof AxiosHeaders
            ? config.headers
            : new AxiosHeaders(config.headers as Record<string, string>);
        headers.set('Authorization', `Bearer ${token}`);
        config.headers = headers;
    }
    return config;
});

api.interceptors.response.use(
    response => response,
    error => {
        if (error.response?.status === 401) {
            useAuth.getState().logout();
            return Promise.reject(error);
        }
        return Promise.reject(error);
    }
);

export { api };
