import React from 'react';

export default function SkeletonCard() {
    return (
        <div className="p-6 bg-white dark:bg-gray-800 rounded-2xl shadow animate-pulse">
            <div className="h-6 bg-gray-300 dark:bg-gray-600 rounded w-3/4 mb-4"></div>
            <div className="space-y-2">
                <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-full"></div>
                <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-5/6"></div>
                <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-2/3"></div>
            </div>
            <div className="mt-4 h-4 bg-gray-300 dark:bg-gray-600 rounded w-1/4"></div>
        </div>
    );
}
