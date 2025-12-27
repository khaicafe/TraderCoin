'use client';

import {useState, useEffect} from 'react';
import {toast} from 'react-hot-toast';
import userService, {User} from '@/services/userService';
import {
  UserIcon,
  KeyIcon,
  EnvelopeIcon,
  PhoneIcon,
  LockClosedIcon,
  CheckCircleIcon,
  XCircleIcon,
  QuestionMarkCircleIcon,
} from '@heroicons/react/24/outline';

export default function SettingsPage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activeTab, setActiveTab] = useState<'profile' | 'password'>('profile');
  const [showChatIDGuide, setShowChatIDGuide] = useState(false);

  // Form states cho Profile
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    full_name: '',
    phone: '',
    chat_id: '',
  });

  // Form states cho Change Password
  const [passwordData, setPasswordData] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  });

  useEffect(() => {
    fetchUserProfile();
  }, []);

  const fetchUserProfile = async () => {
    try {
      const data = await userService.getProfile();
      setUser(data);
      console.log('Fetched user profile:', data);
      setFormData({
        username: data.username || '',
        email: data.email || '',
        full_name: data.full_name || '',
        phone: data.phone || '',
        chat_id: data.chat_id || '',
      });
    } catch (error: any) {
      console.error('Error fetching profile:', error);
      toast.error('Không thể tải thông tin user');
    } finally {
      setLoading(false);
    }
  };

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);

    try {
      const updatedUser = await userService.updateProfile(formData);
      setUser(updatedUser);
      toast.success('✅ Cập nhật thông tin thành công!');
    } catch (error: any) {
      console.error('Error updating profile:', error);
      const errorMsg = error.response?.data?.error || 'Lỗi cập nhật thông tin';
      toast.error(`❌ ${errorMsg}`);
    } finally {
      setSaving(false);
    }
  };

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();

    if (passwordData.new_password !== passwordData.confirm_password) {
      toast.error('❌ Mật khẩu mới không khớp!');
      return;
    }

    if (passwordData.new_password.length < 6) {
      toast.error('❌ Mật khẩu phải có ít nhất 6 ký tự!');
      return;
    }

    setSaving(true);

    try {
      await userService.changePassword({
        current_password: passwordData.current_password,
        new_password: passwordData.new_password,
      });

      toast.success('✅ Đổi mật khẩu thành công!');
      setPasswordData({
        current_password: '',
        new_password: '',
        confirm_password: '',
      });
    } catch (error: any) {
      console.error('Error changing password:', error);
      const errorMsg = error.response?.data?.error || 'Lỗi đổi mật khẩu';
      toast.error(`❌ ${errorMsg}`);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 p-8">
        <div className="max-w-4xl mx-auto">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-300 rounded w-1/4 mb-8"></div>
            <div className="bg-white rounded-xl shadow-sm p-8">
              <div className="space-y-4">
                <div className="h-10 bg-gray-200 rounded"></div>
                <div className="h-10 bg-gray-200 rounded"></div>
                <div className="h-10 bg-gray-200 rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Cài Đặt</h1>
        <p className="text-gray-600">Quản lý thông tin tài khoản và bảo mật</p>
      </div>

      {/* User Info Card */}
      <div className="bg-white rounded-xl shadow-sm p-6 mb-6 border border-gray-100">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-full flex items-center justify-center">
              <UserIcon className="w-8 h-8 text-white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-gray-900">
                {user?.full_name || user?.username}
              </h2>
              <p className="text-gray-500">{user?.email}</p>
            </div>
          </div>
          <div className="text-right">
            <div className="flex items-center gap-2">
              {user?.is_active ? (
                <>
                  <CheckCircleIcon className="w-5 h-5 text-green-500" />
                  <span className="text-sm font-medium text-green-600">
                    Active
                  </span>
                </>
              ) : (
                <>
                  <XCircleIcon className="w-5 h-5 text-red-500" />
                  <span className="text-sm font-medium text-red-600">
                    Inactive
                  </span>
                </>
              )}
            </div>
            <p className="text-xs text-gray-400 mt-1">
              Tham gia:{' '}
              {new Date(user?.created_at || '').toLocaleDateString('vi-VN')}
            </p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="border-b border-gray-200">
          <div className="flex">
            <button
              onClick={() => setActiveTab('profile')}
              className={`flex-1 px-6 py-4 text-sm font-medium transition-colors ${
                activeTab === 'profile'
                  ? 'text-indigo-600 border-b-2 border-indigo-600 bg-indigo-50/50'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
              }`}>
              <UserIcon className="w-5 h-5 inline-block mr-2" />
              Thông Tin Cá Nhân
            </button>
            <button
              onClick={() => setActiveTab('password')}
              className={`flex-1 px-6 py-4 text-sm font-medium transition-colors ${
                activeTab === 'password'
                  ? 'text-indigo-600 border-b-2 border-indigo-600 bg-indigo-50/50'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
              }`}>
              <LockClosedIcon className="w-5 h-5 inline-block mr-2" />
              Đổi Mật Khẩu
            </button>
          </div>
        </div>

        <div className="p-8">
          {/* Profile Tab */}
          {activeTab === 'profile' && (
            <form onSubmit={handleProfileUpdate} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Username */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    <KeyIcon className="w-4 h-4 inline-block mr-1" />
                    Tên đăng nhập
                  </label>
                  <input
                    type="text"
                    value={formData.username}
                    onChange={(e) =>
                      setFormData({...formData, username: e.target.value})
                    }
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                    placeholder="username"
                    required
                  />
                </div>

                {/* Email */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    <EnvelopeIcon className="w-4 h-4 inline-block mr-1" />
                    Email
                  </label>
                  <input
                    type="email"
                    value={formData.email}
                    onChange={(e) =>
                      setFormData({...formData, email: e.target.value})
                    }
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                    placeholder="email@example.com"
                    required
                  />
                </div>

                {/* Full Name */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    <UserIcon className="w-4 h-4 inline-block mr-1" />
                    Họ và tên
                  </label>
                  <input
                    type="text"
                    value={formData.full_name}
                    onChange={(e) =>
                      setFormData({...formData, full_name: e.target.value})
                    }
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                    placeholder="Nguyễn Văn A"
                  />
                </div>

                {/* Phone */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    <PhoneIcon className="w-4 h-4 inline-block mr-1" />
                    Số điện thoại
                  </label>
                  <input
                    type="tel"
                    value={formData.phone}
                    onChange={(e) =>
                      setFormData({...formData, phone: e.target.value})
                    }
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                    placeholder="0912345678"
                  />
                </div>

                {/* Chat ID */}
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    <PhoneIcon className="w-4 h-4 inline-block mr-1" />
                    Telegram Chat ID
                    <button
                      type="button"
                      onClick={() => setShowChatIDGuide(true)}
                      className="ml-2 inline-flex items-center gap-1 px-3 py-1 bg-gradient-to-r from-indigo-500 to-purple-500 text-white text-xs font-semibold rounded-full hover:from-indigo-600 hover:to-purple-600 transition-all shadow-md hover:shadow-lg hover:scale-105 animate-pulse">
                      <QuestionMarkCircleIcon className="w-4 h-4" />
                      <span>Làm sao lấy Chat ID?</span>
                    </button>
                  </label>
                  <input
                    type="text"
                    value={formData.chat_id}
                    onChange={(e) =>
                      setFormData({...formData, chat_id: e.target.value})
                    }
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                    placeholder="123456789"
                  />
                  <p className="text-xs text-gray-500 mt-2 flex items-center gap-1">
                    <span className="text-base">�</span>
                    <span>
                      Không biết Chat ID là gì? Click nút{' '}
                      <strong className="text-indigo-600">
                        "Làm sao lấy Chat ID?"
                      </strong>{' '}
                      để xem hướng dẫn chi tiết
                    </span>
                  </p>
                </div>
              </div>

              <div className="flex justify-end pt-4">
                `
                <button
                  type="submit"
                  disabled={saving}
                  className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-medium rounded-lg hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed">
                  {saving ? (
                    <>
                      <svg
                        className="animate-spin -ml-1 mr-2 h-4 w-4 text-white inline-block"
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
                      Đang lưu...
                    </>
                  ) : (
                    'Lưu thay đổi'
                  )}
                </button>
              </div>
            </form>
          )}

          {/* Password Tab */}
          {activeTab === 'password' && (
            <form
              onSubmit={handlePasswordChange}
              className="space-y-6 max-w-xl">
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
                <p className="text-sm text-yellow-800">
                  ⚠️ Mật khẩu phải có ít nhất 6 ký tự. Sau khi đổi mật khẩu, bạn
                  sẽ cần đăng nhập lại.
                </p>
              </div>

              {/* Current Password */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <LockClosedIcon className="w-4 h-4 inline-block mr-1" />
                  Mật khẩu hiện tại
                </label>
                <input
                  type="password"
                  value={passwordData.current_password}
                  onChange={(e) =>
                    setPasswordData({
                      ...passwordData,
                      current_password: e.target.value,
                    })
                  }
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                  placeholder="••••••••"
                  required
                />
              </div>

              {/* New Password */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <LockClosedIcon className="w-4 h-4 inline-block mr-1" />
                  Mật khẩu mới
                </label>
                <input
                  type="password"
                  value={passwordData.new_password}
                  onChange={(e) =>
                    setPasswordData({
                      ...passwordData,
                      new_password: e.target.value,
                    })
                  }
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                  placeholder="••••••••"
                  required
                  minLength={6}
                />
              </div>

              {/* Confirm Password */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <LockClosedIcon className="w-4 h-4 inline-block mr-1" />
                  Xác nhận mật khẩu mới
                </label>
                <input
                  type="password"
                  value={passwordData.confirm_password}
                  onChange={(e) =>
                    setPasswordData({
                      ...passwordData,
                      confirm_password: e.target.value,
                    })
                  }
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
                  placeholder="••••••••"
                  required
                  minLength={6}
                />
                {passwordData.confirm_password &&
                  passwordData.new_password !==
                    passwordData.confirm_password && (
                    <p className="text-red-500 text-sm mt-1">
                      ❌ Mật khẩu không khớp
                    </p>
                  )}
              </div>

              <div className="flex justify-end pt-4">
                <button
                  type="submit"
                  disabled={
                    saving ||
                    passwordData.new_password !== passwordData.confirm_password
                  }
                  className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-medium rounded-lg hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed">
                  {saving ? (
                    <>
                      <svg
                        className="animate-spin -ml-1 mr-2 h-4 w-4 text-white inline-block"
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
                      Đang cập nhật...
                    </>
                  ) : (
                    'Đổi mật khẩu'
                  )}
                </button>
              </div>
            </form>
          )}
        </div>
      </div>

      {/* Chat ID Guide Modal */}
      {showChatIDGuide && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-2xl max-w-lg w-full p-8 relative animate-fade-in">
            {/* Close Button */}
            <button
              onClick={() => setShowChatIDGuide(false)}
              className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 transition-colors">
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>

            {/* Header */}
            <div className="text-center mb-6">
              <div className="bg-indigo-100 rounded-full w-16 h-16 flex items-center justify-center mx-auto mb-4">
                <QuestionMarkCircleIcon className="w-10 h-10 text-indigo-600" />
              </div>
              <h3 className="text-2xl font-bold text-gray-900">
                🔍 Cách lấy Telegram Chat ID
              </h3>
              <p className="text-gray-600 mt-2">
                Làm theo các bước đơn giản sau
              </p>
            </div>

            {/* Steps */}
            <div className="space-y-5">
              {/* Step 1 */}
              <div className="flex items-start gap-4 p-4 bg-blue-50 rounded-lg border border-blue-200">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-600 text-white rounded-full flex items-center justify-center font-bold">
                  1
                </div>
                <div>
                  <h4 className="font-semibold text-gray-900 mb-1">
                    Mở Telegram và tìm kiếm bot
                  </h4>
                  <p className="text-gray-700">
                    Tìm kiếm{' '}
                    <strong className="text-blue-600">@userinfobot</strong>{' '}
                    trong ô tìm kiếm của Telegram
                  </p>
                </div>
              </div>

              {/* Step 2 */}
              <div className="flex items-start gap-4 p-4 bg-green-50 rounded-lg border border-green-200">
                <div className="flex-shrink-0 w-8 h-8 bg-green-600 text-white rounded-full flex items-center justify-center font-bold">
                  2
                </div>
                <div>
                  <h4 className="font-semibold text-gray-900 mb-1">
                    Nhấn Start hoặc gửi tin nhắn
                  </h4>
                  <p className="text-gray-700">
                    Nhấn nút <strong className="text-green-600">/start</strong>{' '}
                    hoặc gửi bất kỳ tin nhắn nào cho bot
                  </p>
                </div>
              </div>

              {/* Step 3 */}
              <div className="flex items-start gap-4 p-4 bg-purple-50 rounded-lg border border-purple-200">
                <div className="flex-shrink-0 w-8 h-8 bg-purple-600 text-white rounded-full flex items-center justify-center font-bold">
                  3
                </div>
                <div>
                  <h4 className="font-semibold text-gray-900 mb-1">
                    Nhận Chat ID của bạn
                  </h4>
                  <p className="text-gray-700">
                    Bot sẽ trả về thông tin cá nhân bao gồm{' '}
                    <strong className="text-purple-600">Chat ID</strong> của bạn
                  </p>
                  <div className="mt-2 p-2 bg-white rounded border border-purple-300 text-sm font-mono text-gray-800">
                    💡 Id:{' '}
                    <span className="text-purple-600 font-bold">123456789</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className="mt-6 p-4 bg-yellow-50 rounded-lg border border-yellow-200">
              <p className="text-sm text-gray-700 flex items-start gap-2">
                <span className="text-yellow-600 text-lg">⚠️</span>
                <span>
                  <strong>Lưu ý:</strong> Chat ID là dãy số duy nhất định danh
                  tài khoản Telegram của bạn. Hãy sao chép và dán vào ô trên để
                  nhận thông báo từ bot.
                </span>
              </p>
            </div>

            {/* Close Button */}
            <div className="mt-6 text-center">
              <button
                onClick={() => setShowChatIDGuide(false)}
                className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-semibold shadow-md hover:shadow-lg">
                Đã hiểu
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
