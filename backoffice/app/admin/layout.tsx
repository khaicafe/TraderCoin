import AdminSidebar from '../../components/AdminSidebar';

export default function AdminLayout({children}: {children: React.ReactNode}) {
  return (
    <div className="flex min-h-screen bg-gray-100">
      <AdminSidebar />
      <main className="flex-1 ml-64 p-8">{children}</main>
    </div>
  );
}
