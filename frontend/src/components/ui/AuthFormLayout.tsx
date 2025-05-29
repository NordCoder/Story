'use client';

import React, {PropsWithChildren, FormEventHandler} from 'react';
import {Button} from '@/components/ui/button';

interface AuthFormLayoutProps {
    title: string;
    children: React.ReactNode;
    isSubmitting: boolean;
    submitLabel: string;
    submittingLabel: string;
    footerText: string;
    footerLink: { href: string; label: string };
    onSubmit: FormEventHandler<HTMLFormElement>;
}

export function AuthFormLayout({
                                   title,
                                   children,
                                   isSubmitting,
                                   submitLabel,
                                   submittingLabel,
                                   footerText,
                                   footerLink,
                                   onSubmit
                               }: PropsWithChildren<AuthFormLayoutProps>) {
    return (
        <div className="flex items-center justify-center min-h-screen bg-gray-50 dark:bg-gray-900">
            <form
                onSubmit={onSubmit}
                className={
                    `
          w-[24rem]                /* фиксированная ширина */
          h-[30rem]                /* фиксированная высота */
          p-8 
          bg-white dark:bg-gray-800 
          rounded-2xl shadow-lg 
          grid grid-rows-[auto,1fr,auto] 
          gap-y-4
        `
                }
            >
                {/* Заголовок */}
                <h1 className="text-2xl font-semibold text-center text-gray-900 dark:text-gray-50">
                    {title}
                </h1>

                {/* Блок полей */}
                <div className="flex flex-col justify-center space-y-4 w-full">
                    {children}
                </div>

                {/* Футер */}
                <div>
                    <Button type="submit" disabled={isSubmitting} className="w-full mb-4">
                        {isSubmitting ? submittingLabel : submitLabel}
                    </Button>
                    <p className="text-center text-gray-600">
                        {footerText}{' '}
                        <a href={footerLink.href} className="font-medium text-blue-600 hover:underline">
                            {footerLink.label}
                        </a>
                    </p>
                </div>
            </form>
        </div>
    );
}
