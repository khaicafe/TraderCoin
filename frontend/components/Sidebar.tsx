'use client';

import Link from 'next/link';
import {usePathname} from 'next/navigation';
import {
  HomeIcon,
  CogIcon,
  ChartBarIcon,
  ClipboardDocumentListIcon,
  ChartBarSquareIcon,
  DocumentTextIcon,
  ArrowRightOnRectangleIcon,
  KeyIcon,
  CpuChipIcon,
  BellAlertIcon,
  Cog6ToothIcon,
} from '@heroicons/react/24/outline';

const navigation = [
  {name: 'Dashboard', href: '/dashboard', icon: HomeIcon},
  // {name: 'Exchange Keys', href: '/exchange-keys', icon: KeyIcon},
  {name: 'Bot Configs', href: '/bot-configs', icon: CogIcon},
  {name: 'Đặt Lệnh', href: '/trading', icon: ChartBarIcon},
  {name: 'Signals', href: '/signals', icon: BellAlertIcon},
  {
    name: 'Monitoring (Orders)',
    href: '/orders',
    icon: ClipboardDocumentListIcon,
  },
  // {name: 'Monitoring', href: '/monitoring', icon: ChartBarSquareIcon},
  {name: 'Nhật Ký / Lỗi', href: '/logs', icon: DocumentTextIcon},
  {name: 'Cài Đặt', href: '/settings', icon: Cog6ToothIcon},
];

export default function Sidebar() {
  const pathname = usePathname();

  const handleLogout = () => {
    localStorage.removeItem('token');
    window.location.href = '/login';
  };

  return (
    <div className="flex flex-col h-screen w-64 bg-gradient-to-b from-indigo-600 via-purple-600 to-purple-700 text-white shadow-2xl">
      {/* Logo */}
      <div className="flex items-center px-6 h-20 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-white/20 rounded-lg flex items-center justify-center backdrop-blur-sm">
            <CpuChipIcon className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-xl font-bold text-white">Trading Bot</h1>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-6 space-y-1 overflow-y-auto">
        {navigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={`flex items-center px-4 py-3 rounded-lg transition-all duration-200 ${
                isActive
                  ? 'bg-white/20 text-white backdrop-blur-sm shadow-lg'
                  : 'text-white/90 hover:bg-white/10 hover:text-white'
              }`}>
              <item.icon className="w-5 h-5 mr-3" />
              <span className="text-sm font-medium">{item.name}</span>
            </Link>
          );
        })}
      </nav>

      {/* User Section */}
      <div className="border-t border-white/10 p-6 space-y-3">
        <div className="text-center">
          <p className="text-sm text-white/80 mb-3">
            Đang nhập:{' '}
            <span className="font-semibold text-white">bypass-user</span>
          </p>
          <button
            onClick={handleLogout}
            className="w-full px-4 py-3 text-sm font-medium text-white bg-white/10 hover:bg-white/20 rounded-lg transition-all duration-200 border border-white/20">
            Đăng xuất
          </button>
        </div>
      </div>
    </div>
  );
}
