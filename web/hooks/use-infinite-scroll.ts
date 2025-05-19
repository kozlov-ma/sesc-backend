import { useEffect, useRef, RefObject } from 'react';

interface UseInfiniteScrollOptions {
  onLoadMore: () => void;
  threshold?: number;
}

export function useInfiniteScroll({ onLoadMore, threshold = 100 }: UseInfiniteScrollOptions): { ref: RefObject<HTMLDivElement | null> } {
  const observer = useRef<IntersectionObserver | null>(null);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    observer.current = new IntersectionObserver(
      (entries) => {
        const first = entries[0];
        if (first.isIntersecting) {
          onLoadMore();
        }
      },
      {
        rootMargin: `${threshold}px`,
      }
    );

    observer.current.observe(element);

    return () => {
      if (observer.current) {
        observer.current.disconnect();
      }
    };
  }, [onLoadMore, threshold]);

  return { ref };
} 