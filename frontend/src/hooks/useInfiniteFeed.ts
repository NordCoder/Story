import { useInfiniteQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Article } from '@/types/Article';

interface FeedResponse {
    article: Article;
}

export function useInfiniteFeed() {
    return useInfiniteQuery<FeedResponse, Error>({
        queryKey: ['feed'],
        queryFn: async ({ pageParam = 0 }) => {
            const { data } = await api.get<FeedResponse>(`/story/fact`);
            return data;
        },
        getNextPageParam: (_last, pages) => pages.length,
        initialPageParam: 0,
        staleTime: 5 * 60 * 1000,
        retry: 1,
    });
}

