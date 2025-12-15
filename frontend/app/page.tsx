'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    // Check if user is logged in
    const token = localStorage.getItem('token');
    
    if (token) {
      // If logged in, redirect to dashboard
      router.push('/dashboard');
    } else {
      // If not logged in, redirect to login
      router.push('/login');
    }
  }, [router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-600 to-purple-700">
      <div className="text-center">
        <div className="animate-spin rounded-full h-16 w-16 border-t-4 border-white mx-auto mb-4"></div>
        <p className="text-white text-xl font-semibold">Loading TraderCoin...</p>
      </div>
    </div>
  );
}
