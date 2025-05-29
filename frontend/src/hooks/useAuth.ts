import {create} from 'zustand';
import {persist, createJSONStorage} from 'zustand/middleware';

/**
 * Auth state and actions for login/logout and guest mode.
 */
export interface AuthState {
    accessToken: string | null;
    setToken: (token: string) => void;
    logout: () => void;
}

/**
 * Zustand store with persistence in sessionStorage.
 * Use this hook to access and modify the authentication state.
 */
export const useAuth = create<AuthState>()(
    persist(
        (set) => ({
            accessToken: null,
            setToken: (token) => set({accessToken: token}),
            logout: () => set({accessToken: null}),
        }),
        {
            name: 'auth-storage',
            storage: createJSONStorage(() => sessionStorage),
        }
    )
);
