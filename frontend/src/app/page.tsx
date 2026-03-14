'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';

export default function Home() {
  const router = useRouter();
  const pathname = usePathname();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading && pathname === '/') {
      if (isAuthenticated) {
        router.replace('/tokens');
      } else {
        router.replace('/login');
      }
    }
  }, [isAuthenticated, isLoading, router, pathname]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-muted-foreground">加载中...</div>
    </div>
  );
}
