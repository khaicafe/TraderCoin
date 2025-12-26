'use client';

import {useState} from 'react';
import {InformationCircleIcon, XMarkIcon} from '@heroicons/react/24/outline';

export default function TPTrailingStopGuide() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      {/* Nút mở hướng dẫn */}
      <button
        type="button"
        onClick={() => setIsOpen(true)}
        className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300">
        <InformationCircleIcon className="w-5 h-5" />
        <span>Hướng dẫn TP + Trailing Stop</span>
      </button>

      {/* Modal */}
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
          <div className="relative w-full max-w-3xl max-h-[90vh] overflow-y-auto bg-white dark:bg-gray-800 rounded-lg shadow-xl">
            {/* Header */}
            <div className="sticky top-0 flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
              <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                🎯 Cơ chế xử lý lệnh trên Binance Futures
              </h3>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700">
                <XMarkIcon className="w-6 h-6" />
              </button>
            </div>

            {/* Content */}
            <div className="p-6 space-y-6">
              {/* Giới thiệu */}
              <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                <p className="text-sm text-gray-700 dark:text-gray-300">
                  <strong>Nguyên tắc OCO (One Cancels Others):</strong> Khi có
                  nhiều lệnh đóng vị thế, lệnh nào khớp TRƯỚC → đóng vị thế →
                  các lệnh còn lại bị{' '}
                  <span className="font-bold text-red-600">HỦY TỰ ĐỘNG</span>
                </p>
              </div>

              {/* Ví dụ minh họa */}
              <div className="space-y-4">
                <h4 className="font-semibold text-gray-900 dark:text-white">
                  📊 Ví dụ với 3 lệnh đóng vị thế:
                </h4>
                <div className="grid grid-cols-3 gap-3">
                  <div className="p-3 bg-green-50 dark:bg-green-900/20 rounded border border-green-200 dark:border-green-700">
                    <div className="text-green-600 dark:text-green-400 font-semibold">
                      ✅ Take Profit
                    </div>
                    <div className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                      Chốt lời cố định
                    </div>
                  </div>
                  <div className="p-3 bg-red-50 dark:bg-red-900/20 rounded border border-red-200 dark:border-red-700">
                    <div className="text-red-600 dark:text-red-400 font-semibold">
                      ✅ Stop Loss
                    </div>
                    <div className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                      Cắt lỗ
                    </div>
                  </div>
                  <div className="p-3 bg-purple-50 dark:bg-purple-900/20 rounded border border-purple-200 dark:border-purple-700">
                    <div className="text-purple-600 dark:text-purple-400 font-semibold">
                      ✅ Trailing Stop
                    </div>
                    <div className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                      Bám theo giá
                    </div>
                  </div>
                </div>
              </div>

              {/* Kịch bản */}
              <div className="space-y-4">
                <h4 className="font-semibold text-gray-900 dark:text-white">
                  📌 Các kịch bản thực tế
                </h4>

                {/* Kịch bản 1 */}
                <div className="p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg space-y-2">
                  <div className="flex items-center gap-2">
                    <span className="text-2xl">🟢</span>
                    <h5 className="font-semibold text-gray-900 dark:text-white">
                      Kịch bản 1: Giá chạy thẳng lên TP
                    </h5>
                  </div>
                  <div className="pl-8 space-y-1 text-sm">
                    <p className="text-gray-700 dark:text-gray-300">
                      📈 LONG: Entry 100 → TP 105 → Trailing (Activation 102,
                      Callback 1%)
                    </p>
                    <p className="text-gray-600 dark:text-gray-400">
                      ➡️ Giá chạy 100 → 105
                    </p>
                    <p className="text-green-600 dark:text-green-400 font-semibold">
                      ✅ TP khớp → Trailing Stop bị hủy
                    </p>
                    <p className="text-xs text-gray-500">
                      ⭐ TP phát huy tác dụng đầy đủ
                    </p>
                  </div>
                </div>

                {/* Kịch bản 2 */}
                <div className="p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg space-y-2">
                  <div className="flex items-center gap-2">
                    <span className="text-2xl">🟡</span>
                    <h5 className="font-semibold text-gray-900 dark:text-white">
                      Kịch bản 2: Trailing Stop khớp TRƯỚC TP
                    </h5>
                  </div>
                  <div className="pl-8 space-y-1 text-sm">
                    <p className="text-gray-700 dark:text-gray-300">
                      📈 LONG: Entry 100 → TP 105 → Trailing (Activation 102,
                      Callback 1%)
                    </p>
                    <p className="text-gray-600 dark:text-gray-400">
                      ➡️ Giá lên 104 rồi quay đầu
                    </p>
                    <p className="text-orange-600 dark:text-orange-400 font-semibold">
                      ❌ Trailing khớp tại ~103 → TP bị hủy
                    </p>
                    <p className="text-xs text-red-500">
                      ⚠️ TP KHÔNG bao giờ được chạm
                    </p>
                  </div>
                </div>

                {/* Kịch bản 3 */}
                <div className="p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg space-y-2">
                  <div className="flex items-center gap-2">
                    <span className="text-2xl">🔴</span>
                    <h5 className="font-semibold text-gray-900 dark:text-white">
                      Kịch bản 3: Giá rơi thẳng xuống SL
                    </h5>
                  </div>
                  <div className="pl-8 space-y-1 text-sm">
                    <p className="text-gray-600 dark:text-gray-400">
                      ➡️ Giá giảm mạnh
                    </p>
                    <p className="text-red-600 dark:text-red-400 font-semibold">
                      ❌ SL khớp → TP + Trailing Stop bị hủy
                    </p>
                  </div>
                </div>
              </div>

              {/* Khi nào dùng */}
              <div className="space-y-3 p-4 bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20 rounded-lg">
                <h4 className="font-semibold text-gray-900 dark:text-white">
                  ⚔️ TP vs Trailing Stop – dùng sao cho đúng?
                </h4>

                <div className="space-y-3">
                  {/* Cách 1 */}
                  <div className="bg-white dark:bg-gray-800 p-3 rounded border-l-4 border-yellow-500">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xl">🥇</span>
                      <strong className="text-gray-900 dark:text-white">
                        Cách 1: TP xa + Trailing gần (BEST)
                      </strong>
                    </div>
                    <ul className="text-sm text-gray-600 dark:text-gray-400 ml-8 space-y-1">
                      <li>• TP: Rất xa (target lý tưởng, ví dụ +10%)</li>
                      <li>
                        • Trailing: Gần hơn (giữ lợi nhuận thực tế, activation
                        +3%)
                      </li>
                      <li className="text-green-600 dark:text-green-400">
                        → TP = "mơ ước" / Trailing = "thực dụng"
                      </li>
                    </ul>
                  </div>

                  {/* Cách 2 */}
                  <div className="bg-white dark:bg-gray-800 p-3 rounded border-l-4 border-gray-500">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xl">🥈</span>
                      <strong className="text-gray-900 dark:text-white">
                        Cách 2: KHÔNG TP – chỉ Trailing
                      </strong>
                    </div>
                    <ul className="text-sm text-gray-600 dark:text-gray-400 ml-8 space-y-1">
                      <li>• Dùng cho trend mạnh</li>
                      <li>• Để thị trường tự trả lời</li>
                    </ul>
                  </div>

                  {/* Cách 3 */}
                  <div className="bg-white dark:bg-gray-800 p-3 rounded border-l-4 border-orange-500">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xl">🥉</span>
                      <strong className="text-gray-900 dark:text-white">
                        Cách 3: Chia vị thế (xịn nhất)
                      </strong>
                    </div>
                    <ul className="text-sm text-gray-600 dark:text-gray-400 ml-8 space-y-1">
                      <li>• 50% dùng TP cố định</li>
                      <li>• 50% dùng Trailing Stop</li>
                      <li className="text-green-600 dark:text-green-400">
                        → Ăn chắc + ăn dài
                      </li>
                    </ul>
                  </div>
                </div>
              </div>

              {/* Lưu ý quan trọng */}
              <div className="p-4 bg-red-50 dark:bg-red-900/20 rounded-lg border border-red-200 dark:border-red-700">
                <h4 className="font-semibold text-red-900 dark:text-red-300 mb-2">
                  ⚠️ Lưu ý QUAN TRỌNG
                </h4>
                <ul className="text-sm text-gray-700 dark:text-gray-300 space-y-1">
                  <li>• TP/SL đặt trong khung vị thế ≠ lệnh limit thường</li>
                  <li>
                    • Trailing Stop luôn là lệnh <strong>Market</strong>
                  </li>
                  <li>
                    •{' '}
                    <strong className="text-red-600 dark:text-red-400">
                      KHÔNG có chuyện "khớp cả TP và Trailing"
                    </strong>
                  </li>
                  <li>• Lệnh nào khớp trước → lệnh còn lại bị HỦY</li>
                </ul>
              </div>

              {/* Kết luận */}
              <div className="p-4 bg-gradient-to-r from-green-50 to-blue-50 dark:from-green-900/20 dark:to-blue-900/20 rounded-lg">
                <h4 className="font-semibold text-gray-900 dark:text-white mb-2">
                  🧠 KẾT LUẬN
                </h4>
                <p className="text-sm text-gray-700 dark:text-gray-300">
                  TP vẫn có tác dụng, nhưng thường bị Trailing Stop{' '}
                  <strong>"ăn mất"</strong> nếu Trailing chạy sớm. Hãy đặt TP xa
                  hơn Trailing để tối ưu chiến lược!
                </p>
              </div>
            </div>

            {/* Footer */}
            <div className="sticky bottom-0 p-4 border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
              <button
                onClick={() => setIsOpen(false)}
                className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors">
                Đã hiểu
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
