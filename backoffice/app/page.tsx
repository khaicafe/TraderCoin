'use client';

import {useState} from 'react';
import {useRouter} from 'next/navigation';

export default function AdminLoginPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/admin/login`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(formData),
        },
      );

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Login failed');
      }

      // Backend currently returns placeholder response
      // Store token (or mock token if not provided)
      const token = data.token || 'mock-admin-token-' + Date.now();
      localStorage.setItem('admin_token', token);

      // Store admin data
      const adminData = data.admin || {
        id: 1,
        name: 'Admin User',
        email: formData.email,
        role: 'admin',
      };
      localStorage.setItem('admin', JSON.stringify(adminData));

      router.push('/admin/dashboard');
    } catch (err: any) {
      setError(err.message || 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-700 flex items-center justify-center p-4">
      <div className="max-w-md w-full">
        <div className="bg-white rounded-2xl shadow-2xl p-8">
          <div className="text-center mb-8">
            <div className="inline-block p-4 bg-slate-100 rounded-full mb-4">
              <svg
                className="w-12 h-12 text-slate-700"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                />
              </svg>
            </div>
            <h1 className="text-3xl font-bold text-slate-900 mb-2">
              Admin Portal
            </h1>
            <p className="text-gray-600">TraderCoin Backoffice</p>
          </div>

          {error && (
            <div className="mb-4 p-4 bg-red-50 border border-red-200 text-red-600 rounded-lg text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Admin Email
              </label>
              <input
                type="email"
                required
                value={formData.email}
                onChange={(e) =>
                  setFormData({...formData, email: e.target.value})
                }
                className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-slate-600 focus:outline-none text-gray-900 bg-white"
                placeholder="admin@tradercoin.com"
              />
            </div>

            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Password
              </label>
              <input
                type="password"
                required
                value={formData.password}
                onChange={(e) =>
                  setFormData({...formData, password: e.target.value})
                }
                className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg focus:border-slate-600 focus:outline-none text-gray-900 bg-white"
                placeholder="••••••••"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-slate-800 hover:bg-slate-900 text-white font-bold py-3 rounded-lg transition-colors disabled:opacity-50">
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <svg
                    className="animate-spin h-5 w-5"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24">
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Signing in...
                </span>
              ) : (
                'Sign In to Admin Panel'
              )}
            </button>
          </form>

          <div className="mt-6 text-center text-xs text-gray-500">
            <p>🔒 Secure admin access only</p>
          </div>
        </div>

        <div className="text-center mt-6 text-sm text-gray-300">
          <p>© 2025 TraderCoin Backoffice. All rights reserved.</p>
        </div>
      </div>
    </div>
  );
}
