import '@/styles/globals.css';
import { PropsWithChildren } from 'react';
import Providers from '@/components/Providers';
import Header from '@/components/header/Header';

export default function RootLayout({ children }: PropsWithChildren) {
    return (
        <html lang="ru" suppressHydrationWarning>
        <body className="min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-50">
        <Providers>
            <Header />
            <main className="pt-16 h-[calc(100vh-4rem)] flex items-center justify-center overflow-hidden">
                {children}
            </main>
        </Providers>
        </body>
        </html>
    );
}