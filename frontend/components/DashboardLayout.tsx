'use client';

import {usePathname} from 'next/navigation';
import Sidebar from './Sidebar';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();

  // Pages that don't need sidebar
  const publicPages = ['/', '/login', '/register'];
  const showSidebar = !publicPages.includes(pathname);

  if (!showSidebar) {
    return <>{children}</>;
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar />
      <main className="flex-1 overflow-y-auto bg-gray-100">
        <div className="p-8">{children}</div>
      </main>
    </div>
  );
}
