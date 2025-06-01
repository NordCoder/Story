'use client';

import React, { useState, useRef } from 'react';
import { Article } from '@/types/Article';
import { useAuth } from '@/hooks/useAuth';
import { useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';
import { Heart } from 'lucide-react';
import clsx from 'clsx';
import Image from 'next/image';

interface ArticleCardProps {
    article: Article;
}

function paginateText(text: string, charsPerPage = 200): string[] {
    const words = text.split(' ');
    const pages: string[] = [];
    let current = '';

    for (const word of words) {
        if ((current + ' ' + word).trim().length > charsPerPage) {
            pages.push(current.trim());
            current = word;
        } else {
            current = (current + ' ' + word).trim();
        }
    }
    if (current) pages.push(current.trim());
    return pages;
}

export default function ArticleCard({ article }: ArticleCardProps) {
    const { accessToken } = useAuth();
    const [liked, setLiked] = useState(false);

    const toggleMutation = useMutation<void, Error, boolean>({
        mutationFn: async (newLiked) => {
            const endpoint = newLiked ? 'like' : 'unlike';
            await api.post(`/recommendations/${endpoint}`, {
                category: article.category,
            });
        },
        onError: (_, newLiked) => {
            // Откат состояния, если ошибка
            setLiked((prev) => !prev);
            toast.error(newLiked ? 'Не удалось поставить лайк' : 'Не удалось убрать лайк');
        },
        onSuccess: (_, newLiked) => {
            toast.success(newLiked ? 'Понравилось!' : 'Лайк удалён');
        },
    });

    const handleLike = () => {
        if (!accessToken) {
            toast('Войдите, чтобы лайкать', { icon: '🔒' });
            return;
        }
        const newLiked = !liked;
        setLiked(newLiked);
        toggleMutation.mutate(newLiked);
    };

    const isLoading = toggleMutation.status === 'pending';

    const pages = paginateText(article.summary, 200);
    const [pageIndex, setPageIndex] = useState(0);

    const touchStartX = useRef(0);
    const touchEndX = useRef(0);

    const handleTouchStart = (e: React.TouchEvent) => {
        touchStartX.current = e.changedTouches[0].screenX;
    };

    const handleTouchEnd = (e: React.TouchEvent) => {
        touchEndX.current = e.changedTouches[0].screenX;
        const delta = touchEndX.current - touchStartX.current;
        const threshold = 50; // минимальное смещение
        if (delta > threshold && pageIndex > 0) {
            setPageIndex((p) => p - 1);
        } else if (delta < -threshold && pageIndex < pages.length - 1) {
            setPageIndex((p) => p + 1);
        }
    };

    return (
        <div className="max-w-xl mx-auto p-6 bg-white dark:bg-gray-800 rounded-2xl shadow-lg hover:shadow-2xl transition-all duration-200">
            {article.imgUrl && (
                <div className="relative w-full h-48 bg-gray-100 rounded-xl overflow-hidden mb-4">
                    <Image
                        src={article.imgUrl}
                        alt={article.title}
                        fill
                        style={{ objectFit: 'cover' }}
                        placeholder="blur"
                        blurDataURL="/img/placeholder.svg"
                        priority={false}
                    />
                </div>
            )}

            <h2 className="text-2xl font-semibold text-gray-900 dark:text-gray-50 mb-4">
                {article.title}
            </h2>

            {/* Блок текста с фиксированной высотой (~7 строк) и свайпом */}
            <div
                className="relative mb-6 h-28 overflow-hidden"
                onTouchStart={handleTouchStart}
                onTouchEnd={handleTouchEnd}
            >
                <div className="h-full px-2">
                    <p className="text-gray-700 dark:text-gray-300">{pages[pageIndex]}</p>
                </div>

                {/* Стрелки навигации */}
                {pageIndex > 0 && (
                    <button
                        onClick={() => setPageIndex((p) => p - 1)}
                        className="absolute left-1 top-1/2 -translate-y-1/2 bg-white dark:bg-gray-700 bg-opacity-80 p-1 rounded-full shadow"
                    >
                        ◀
                    </button>
                )}
                {pageIndex < pages.length - 1 && (
                    <button
                        onClick={() => setPageIndex((p) => p + 1)}
                        className="absolute right-1 top-1/2 -translate-y-1/2 bg-white dark:bg-gray-700 bg-opacity-80 p-1 rounded-full shadow"
                    >
                        ▶
                    </button>
                )}

                {/* Индикатор страниц */}
                {pages.length > 1 && (
                    <div className="absolute bottom-1 left-1/2 -translate-x-1/2 flex space-x-1">
                        {pages.map((_, idx) => (
                            <div
                                key={idx}
                                className={clsx(
                                    'w-2 h-2 rounded-full',
                                    idx === pageIndex
                                        ? 'bg-gray-900 dark:bg-gray-100'
                                        : 'bg-gray-400 dark:bg-gray-600'
                                )}
                            />
                        ))}
                    </div>
                )}
            </div>

            <div className="flex justify-between items-center">
                <button
                    onClick={handleLike}
                    disabled={isLoading}
                    className={clsx(
                        'flex items-center space-x-2 text-lg',
                        isLoading
                            ? 'opacity-50 cursor-not-allowed text-gray-400'
                            : liked
                                ? 'text-red-500'
                                : 'text-gray-600 hover:text-red-500'
                    )}
                >
                    <Heart size={24} />
                    <span>{liked ? 'Liked' : 'Like'}</span>
                </button>

                <a
                    href={article.wikiUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="text-blue-700 dark:text-blue-400 hover:underline font-medium"
                >
                    Читать оригинал
                </a>
            </div>
        </div>
    );
}