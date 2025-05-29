'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'react-hot-toast';
import ArticleCard from '@/components/feed/ArticleCard';
import { api } from '@/lib/api';

interface Fact {
    title: string;
    category: string;
    summary: string;
    wikiUrl: string;
    imgUrl: string;
}

export default function FactFeed() {
    const router = useRouter();
    const [facts, setFacts] = useState<Fact[]>([]);
    const [loading, setLoading] = useState(false);
    const loaderRef = useRef<HTMLDivElement | null>(null);
    const unauthorizedRef = useRef(false);

    const fetchFact = async () => {
        if (unauthorizedRef.current || loading) return;
        setLoading(true);
        try {
            const { data } = await api.get<{ fact: Fact }>('/story/fact');
            setFacts(prev => [...prev, data.fact]);
        } catch (e: any) {
            if (e.response?.status === 401) {
                unauthorizedRef.current = true;
                toast.error('Пожалуйста, авторизуйтесь');
                router.replace('/login');
            } else {
                console.error('Ошибка при получении факта:', e);
            }
        } finally {
            setLoading(false);
        }
    };



    useEffect(() => {
        let observer: IntersectionObserver;
        const initFetchAndObserve = async () => {
            await fetchFact();
            if (unauthorizedRef.current) return;
            if (!loaderRef.current) return;

            observer = new IntersectionObserver(
                entries => {
                    if (entries[0].isIntersecting) fetchFact();
                },
                { root: null, rootMargin: '200px 0px', threshold: 0 }
            );

            observer.observe(loaderRef.current);
        };

        initFetchAndObserve();

        return () => {
            observer?.disconnect();
        };
    }, []);

    return (
        <div className="w-full h-full overflow-y-auto snap-y snap-mandatory scroll-smooth no-scrollbar">
            {facts.map((fact, idx) => (
                <div
                    key={idx}
                    className="snap-start flex items-center justify-center h-full px-4"
                    style={{ scrollSnapStop: 'always' }}
                >
                    <div className="w-full max-w-md">
                        <ArticleCard article={fact as any} />
                    </div>
                </div>
            ))}

            <div ref={loaderRef} className="h-8 flex items-center justify-center">
                {loading && <span className="text-lg">Загрузка...</span>}
            </div>
        </div>
    );
}
