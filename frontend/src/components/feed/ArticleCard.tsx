'use client';

import React, { useState } from 'react';
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

export default function ArticleCard({ article }: ArticleCardProps) {
    const { accessToken } = useAuth();
    const [liked, setLiked] = useState(false);

    const toggleMutation = useMutation<void, Error, boolean>({
        mutationFn: async (newLiked) => {
            const endpoint = newLiked ? 'like' : 'unlike';
            await api.post(`/recommendations/${endpoint}`, { category: article.category });
        },
        onError: (_, newLiked) => {
            // rollback
            setLiked(prev => !prev);
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

            <h3 className="text-2xl font-semibold text-gray-900 dark:text-gray-50 mb-4">
                {article.title}
            </h3>

            {/* Блок с горизонтальной прокруткой для длинного текста */}
            <div className="mb-6">
                <div className="overflow-x-auto whitespace-nowrap px-2 py-1 bg-gray-50 dark:bg-gray-700 rounded-lg">
                    <p className="inline-block text-gray-700 dark:text-gray-300">
                        {article.summary}
                    </p>
                </div>
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
