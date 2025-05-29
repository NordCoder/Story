"use client";
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { useTheme } from 'next-themes';
import { useEffect, useState } from 'react';
import { Sun, Moon, User, LogIn } from 'lucide-react';

export default function Header() {
    const { accessToken, logout } = useAuth();
    const router = useRouter();
    const { theme, setTheme } = useTheme();
    const [mounted, setMounted] = useState(false);
    useEffect(() => setMounted(true), []);

    const toggleTheme = () => setTheme(theme === 'dark' ? 'light' : 'dark');

    return (
        <header className="fixed top-0 left-0 w-full h-16 bg-white dark:bg-gray-800 shadow-md z-20 px-6 flex items-center justify-between">
            <Link href="/" className="text-2xl font-bold text-blue-600">Story</Link>
            <nav className="flex items-center space-x-4">
                {mounted && (
                    <button onClick={toggleTheme} aria-label="Toggle Theme" className="p-2 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700">
                        {theme === 'dark' ? <Sun size={20} /> : <Moon size={20} />}
                    </button>
                )}
                {accessToken ? (
                    <>
                        <button aria-label="Profile" className="p-2 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700">
                            <User size={20} />
                        </button>
                        <button onClick={() => { logout(); router.push('/login'); }} className="flex items-center space-x-1 px-3 py-1 bg-red-500 text-white rounded-lg hover:bg-red-600">
                            <LogIn size={16} />
                            <span>Logout</span>
                        </button>
                    </>
                ) : (
                    <Link href="/login" className="flex items-center space-x-1 px-3 py-1 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                        <LogIn size={16} />
                        <span>Login</span>
                    </Link>
                )}
            </nav>
        </header>
    );
}