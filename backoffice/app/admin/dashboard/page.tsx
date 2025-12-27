'use client';

import {useEffect, useState} from 'react';
import {useRouter} from 'next/navigation';
import {
  getOrders,
  getUsers,
  getTransactions,
  getSignals,
} from '@/services/adminService';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';

export default function AdminDashboard() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [pnlData, setPnlData] = useState<any[]>([]);
  const [userGrowthData, setUserGrowthData] = useState<any[]>([]);
  const [transactionData, setTransactionData] = useState<any[]>([]);
  const [signalData, setSignalData] = useState<any[]>([]);
  const [totalPnL, setTotalPnL] = useState(0);
  const [totalUsers, setTotalUsers] = useState(0);
  const [totalSignals, setTotalSignals] = useState(0);

  useEffect(() => {
    const token = localStorage.getItem('admin_token');
    if (!token) {
      router.push('/');
      return;
    }
    fetchDashboardData();
  }, [router]);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);

      const ordersResponse = await getOrders();
      const orders = ordersResponse.orders || [];
      const pnlByDay = generatePnLData(orders);
      setPnlData(pnlByDay);
      const total = orders
        .filter((o: any) => o.status === 'filled')
        .reduce((sum: number, order: any) => sum + (order.pnl || 0), 0);
      setTotalPnL(total);

      const usersResponse = await getUsers();
      const users = usersResponse.users || [];
      setTotalUsers(users.length);
      const userGrowth = generateUserGrowthData(users);
      setUserGrowthData(userGrowth);

      const transactions = await getTransactions();
      const transactionByMonth = generateTransactionData(transactions);
      setTransactionData(transactionByMonth);

      const signalsResponse = await getSignals();
      const signals = signalsResponse.signals || [];
      const signalStats = generateSignalData(signals);
      setSignalData(signalStats);
      const totalSig = signalStats.reduce((sum, s) => sum + s.value, 0);
      setTotalSignals(totalSig);
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  const generatePnLData = (orders: any[]) => {
    const last7Days = [];
    for (let i = 6; i >= 0; i--) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      const dateStr = date.toISOString().split('T')[0];
      const dayOrders = orders.filter((order: any) => {
        const orderDate = new Date(order.created_at)
          .toISOString()
          .split('T')[0];
        return orderDate === dateStr && order.status === 'filled';
      });
      const totalPnL = dayOrders.reduce(
        (sum: number, order: any) => sum + (order.pnl || 0),
        0,
      );
      last7Days.push({
        date: date.toLocaleDateString('en-US', {
          month: 'short',
          day: 'numeric',
        }),
        pnl: parseFloat(totalPnL.toFixed(2)),
        profit: totalPnL > 0 ? totalPnL : 0,
        loss: totalPnL < 0 ? Math.abs(totalPnL) : 0,
      });
    }
    return last7Days;
  };

  const generateUserGrowthData = (users: any[]) => {
    const last6Months = [];
    for (let i = 5; i >= 0; i--) {
      const date = new Date();
      date.setMonth(date.getMonth() - i);
      const monthStr = date.toISOString().slice(0, 7);
      const monthUsers = users.filter((user: any) =>
        user.created_at.startsWith(monthStr),
      );
      last6Months.push({
        month: date.toLocaleDateString('en-US', {month: 'short'}),
        total: monthUsers.length,
        active: monthUsers.filter((u: any) => u.status === 'active').length,
        suspended: monthUsers.filter((u: any) => u.status === 'suspended')
          .length,
      });
    }
    return last6Months;
  };

  const generateTransactionData = (transactions: any[]) => {
    const last6Months = [];
    for (let i = 5; i >= 0; i--) {
      const date = new Date();
      date.setMonth(date.getMonth() - i);
      const monthStr = date.toISOString().slice(0, 7);
      const monthTransactions = transactions.filter((tx: any) =>
        tx.created_at.startsWith(monthStr),
      );
      const deposits = monthTransactions
        .filter((tx: any) => tx.type === 'deposit')
        .reduce((sum: number, tx: any) => sum + (tx.amount || 0), 0);
      const withdrawals = monthTransactions
        .filter((tx: any) => tx.type === 'withdrawal')
        .reduce((sum: number, tx: any) => sum + (tx.amount || 0), 0);
      last6Months.push({
        month: date.toLocaleDateString('en-US', {month: 'short'}),
        deposits: parseFloat(deposits.toFixed(2)),
        withdrawals: parseFloat(withdrawals.toFixed(2)),
      });
    }
    return last6Months;
  };

  const generateSignalData = (signals: any[]) => {
    const executed = signals.filter((s: any) => s.status === 'executed').length;
    const pending = signals.filter((s: any) => s.status === 'pending').length;
    const failed = signals.filter(
      (s: any) => s.status === 'failed' || s.status === 'error',
    ).length;
    return [
      {name: 'Executed', value: executed, color: '#10b981'},
      {name: 'Pending', value: pending, color: '#f59e0b'},
      {name: 'Failed', value: failed, color: '#ef4444'},
    ];
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-gray-50 via-orange-50 to-gray-50">
        <div className="text-center">
          <div className="relative">
            <div className="animate-spin rounded-full h-16 w-16 border-4 border-orange-200 mx-auto"></div>
            <div className="animate-spin rounded-full h-16 w-16 border-t-4 border-orange-500 absolute top-0 left-1/2 transform -translate-x-1/2"></div>
          </div>
          <p className="mt-6 text-gray-600 font-medium">
            Loading Analytics Dashboard...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <div className="flex items-center justify-between flex-wrap gap-4">
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-orange-600 via-red-500 to-pink-500 bg-clip-text text-transparent mb-1">
              Analytics Dashboard
            </h1>
            <p className="text-sm text-gray-600 flex items-center gap-2">
              <span className="inline-block w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
              Real-time trading insights & performance metrics
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="bg-white px-4 py-2 rounded-xl shadow-lg border border-orange-100 hover:shadow-xl transition-all">
              <div className="text-xs text-gray-500 mb-1">Total PnL</div>
              <div
                className={`text-xl font-bold ${
                  totalPnL >= 0 ? 'text-green-600' : 'text-red-600'
                }`}>
                ${Math.abs(totalPnL).toFixed(2)}
              </div>
            </div>
            <div className="bg-white px-4 py-2 rounded-xl shadow-lg border border-blue-100 hover:shadow-xl transition-all">
              <div className="text-xs text-gray-500 mb-1">Active Users</div>
              <div className="text-xl font-bold text-blue-600">
                {totalUsers}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* PnL Chart */}
        <div className="group bg-white rounded-2xl shadow-xl hover:shadow-2xl transition-all duration-500 overflow-hidden border border-gray-100 hover:border-orange-200 flex flex-col">
          <div className="bg-gradient-to-br from-orange-500 via-red-500 to-pink-500 p-2 text-white relative overflow-hidden flex-shrink-0">
            <div className="absolute top-0 right-0 w-24 h-24 bg-white/10 rounded-full -mr-12 -mt-12"></div>
            <div className="absolute bottom-0 left-0 w-20 h-20 bg-white/10 rounded-full -ml-10 -mb-10"></div>
            <div className="relative z-10">
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-1.5">
                    <div className="p-1 bg-white/20 backdrop-blur-sm rounded-lg">
                      <svg
                        className="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2.5}
                          d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                        />
                      </svg>
                    </div>
                    <h2 className="text-base font-bold">PnL Performance</h2>
                  </div>
                </div>
                <div className="text-right bg-white/20 backdrop-blur-md px-2.5 py-1 rounded-lg border border-white/30">
                  <div className="text-lg font-bold">
                    ${Math.abs(totalPnL).toFixed(2)}
                  </div>
                  <div className="text-xs flex items-center gap-1 justify-end font-semibold">
                    {totalPnL >= 0 ? (
                      <span className="text-green-200">↑ Profit</span>
                    ) : (
                      <span className="text-red-200">↓ Loss</span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="p-4 bg-gradient-to-b from-white to-gray-50 flex-1 flex items-center">
            <ResponsiveContainer width="100%" height="100%" minHeight={250}>
              <AreaChart data={pnlData}>
                <defs>
                  <linearGradient id="colorProfit" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.9} />
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0.1} />
                  </linearGradient>
                  <linearGradient id="colorLoss" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#ef4444" stopOpacity={0.9} />
                    <stop offset="95%" stopColor="#ef4444" stopOpacity={0.1} />
                  </linearGradient>
                </defs>
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="#e5e7eb"
                  vertical={false}
                  strokeOpacity={0.5}
                />
                <XAxis
                  dataKey="date"
                  stroke="#9ca3af"
                  style={{fontSize: '12px', fontWeight: 600}}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  stroke="#9ca3af"
                  style={{fontSize: '12px', fontWeight: 600}}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1f2937',
                    border: 'none',
                    borderRadius: '12px',
                    boxShadow: '0 20px 40px rgba(0,0,0,0.3)',
                    color: '#fff',
                    padding: '12px',
                    fontWeight: 600,
                  }}
                  formatter={(value: any) => [`$${value.toFixed(2)}`, '']}
                />
                <Area
                  type="monotone"
                  dataKey="profit"
                  stroke="#10b981"
                  strokeWidth={3}
                  fill="url(#colorProfit)"
                  stackId="1"
                  animationDuration={1500}
                />
                <Area
                  type="monotone"
                  dataKey="loss"
                  stroke="#ef4444"
                  strokeWidth={3}
                  fill="url(#colorLoss)"
                  stackId="1"
                  animationDuration={1500}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* User Growth Chart */}
        <div className="group bg-white rounded-2xl shadow-xl hover:shadow-2xl transition-all duration-500 overflow-hidden border border-gray-100 hover:border-blue-200 flex flex-col">
          <div className="bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-600 p-2 text-white relative overflow-hidden flex-shrink-0">
            <div className="absolute top-0 right-0 w-24 h-24 bg-white/10 rounded-full -mr-12 -mt-12"></div>
            <div className="absolute bottom-0 left-0 w-20 h-20 bg-white/10 rounded-full -ml-10 -mb-10"></div>
            <div className="relative z-10">
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-1.5">
                    <div className="p-1 bg-white/20 backdrop-blur-sm rounded-lg">
                      <svg
                        className="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2.5}
                          d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                        />
                      </svg>
                    </div>
                    <h2 className="text-base font-bold">User Growth</h2>
                  </div>
                </div>
                <div className="text-right bg-white/20 backdrop-blur-md px-2.5 py-1 rounded-lg border border-white/30">
                  <div className="text-lg font-bold">{totalUsers}</div>
                  <div className="text-xs font-semibold">Total Users</div>
                </div>
              </div>
            </div>
          </div>
          <div className="p-4 bg-gradient-to-b from-white to-gray-50 flex-1 flex items-center">
            <ResponsiveContainer width="100%" height="100%" minHeight={250}>
              <BarChart data={userGrowthData} barGap={8}>
                <defs>
                  <linearGradient id="colorActive" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#3b82f6" stopOpacity={1} />
                    <stop offset="100%" stopColor="#60a5fa" stopOpacity={0.8} />
                  </linearGradient>
                  <linearGradient
                    id="colorSuspended"
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1">
                    <stop offset="0%" stopColor="#ef4444" stopOpacity={1} />
                    <stop offset="100%" stopColor="#f87171" stopOpacity={0.8} />
                  </linearGradient>
                </defs>
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="#e5e7eb"
                  vertical={false}
                  strokeOpacity={0.5}
                />
                <XAxis
                  dataKey="month"
                  stroke="#9ca3af"
                  style={{fontSize: '12px', fontWeight: 600}}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  stroke="#9ca3af"
                  style={{fontSize: '12px', fontWeight: 600}}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1f2937',
                    border: 'none',
                    borderRadius: '12px',
                    boxShadow: '0 20px 40px rgba(0,0,0,0.3)',
                    color: '#fff',
                    padding: '12px',
                    fontWeight: 600,
                  }}
                />
                <Legend
                  wrapperStyle={{paddingTop: '20px', fontWeight: 600}}
                  iconType="circle"
                />
                <Bar
                  dataKey="active"
                  fill="url(#colorActive)"
                  name="Active Users"
                  radius={[8, 8, 0, 0]}
                  animationDuration={1500}
                />
                <Bar
                  dataKey="suspended"
                  fill="url(#colorSuspended)"
                  name="Suspended"
                  radius={[8, 8, 0, 0]}
                  animationDuration={1500}
                />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Transactions Chart */}
        <div className="group bg-white rounded-2xl shadow-xl hover:shadow-2xl transition-all duration-500 overflow-hidden border border-gray-100 hover:border-green-200 flex flex-col">
          <div className="bg-gradient-to-br from-green-500 via-emerald-500 to-teal-600 p-2 text-white relative overflow-hidden flex-shrink-0">
            <div className="absolute top-0 right-0 w-24 h-24 bg-white/10 rounded-full -mr-12 -mt-12"></div>
            <div className="absolute bottom-0 left-0 w-20 h-20 bg-white/10 rounded-full -ml-10 -mb-10"></div>
            <div className="relative z-10">
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-1.5">
                    <div className="p-1 bg-white/20 backdrop-blur-sm rounded-lg">
                      <svg
                        className="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2.5}
                          d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    </div>
                    <h2 className="text-base font-bold">Transactions</h2>
                  </div>
                </div>
                <div className="text-right bg-white/20 backdrop-blur-md px-2.5 py-1 rounded-lg border border-white/30">
                  <div className="text-lg font-bold">
                    $
                    {transactionData
                      .reduce((sum, t) => sum + t.deposits - t.withdrawals, 0)
                      .toLocaleString()}
                  </div>
                  <div className="text-xs font-semibold">Net Balance</div>
                </div>
              </div>
            </div>
          </div>
          <div className="p-4 bg-gradient-to-b from-white to-gray-50 flex-1 flex items-center">
            <ResponsiveContainer width="100%" height="100%" minHeight={250}>
              <LineChart data={transactionData}>
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="#e5e7eb"
                  vertical={false}
                  strokeOpacity={0.5}
                />
                <XAxis
                  dataKey="month"
                  stroke="#9ca3af"
                  style={{fontSize: '12px', fontWeight: 600}}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  stroke="#9ca3af"
                  style={{fontSize: '12px', fontWeight: 600}}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1f2937',
                    border: 'none',
                    borderRadius: '12px',
                    boxShadow: '0 20px 40px rgba(0,0,0,0.3)',
                    color: '#fff',
                    padding: '12px',
                    fontWeight: 600,
                  }}
                  formatter={(value: any) => [`$${value.toLocaleString()}`, '']}
                />
                <Legend
                  wrapperStyle={{paddingTop: '20px', fontWeight: 600}}
                  iconType="circle"
                />
                <Line
                  type="monotone"
                  dataKey="deposits"
                  stroke="#10b981"
                  strokeWidth={4}
                  name="Deposits"
                  dot={{r: 6, fill: '#10b981', strokeWidth: 3, stroke: '#fff'}}
                  activeDot={{r: 8, fill: '#10b981', strokeWidth: 0}}
                  animationDuration={1500}
                />
                <Line
                  type="monotone"
                  dataKey="withdrawals"
                  stroke="#ef4444"
                  strokeWidth={4}
                  name="Withdrawals"
                  dot={{r: 6, fill: '#ef4444', strokeWidth: 3, stroke: '#fff'}}
                  activeDot={{r: 8, fill: '#ef4444', strokeWidth: 0}}
                  animationDuration={1500}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Trading Signals Chart */}
        <div className="group bg-white rounded-2xl shadow-xl hover:shadow-2xl transition-all duration-500 overflow-hidden border border-gray-100 hover:border-purple-200 flex flex-col">
          <div className="bg-gradient-to-br from-purple-500 via-pink-500 to-rose-500 p-2 text-white relative overflow-hidden flex-shrink-0">
            <div className="absolute top-0 right-0 w-24 h-24 bg-white/10 rounded-full -mr-12 -mt-12"></div>
            <div className="absolute bottom-0 left-0 w-20 h-20 bg-white/10 rounded-full -ml-10 -mb-10"></div>
            <div className="relative z-10">
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-1.5">
                    <div className="p-1 bg-white/20 backdrop-blur-sm rounded-lg">
                      <svg
                        className="w-3.5 h-3.5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24">
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2.5}
                          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
                        />
                      </svg>
                    </div>
                    <h2 className="text-base font-bold">Trading Signals</h2>
                  </div>
                </div>
                <div className="text-right bg-white/20 backdrop-blur-md px-2.5 py-1 rounded-lg border border-white/30">
                  <div className="text-lg font-bold">{totalSignals}</div>
                  <div className="text-xs font-semibold">Total Signals</div>
                </div>
              </div>
            </div>
          </div>
          <div className="p-4 bg-gradient-to-b from-white to-gray-50 flex-1 flex items-center">
            <ResponsiveContainer width="100%" height="100%" minHeight={250}>
              <PieChart>
                <defs>
                  {signalData.map((entry, index) => (
                    <linearGradient
                      key={index}
                      id={`gradient${index}`}
                      x1="0"
                      y1="0"
                      x2="1"
                      y2="1">
                      <stop
                        offset="0%"
                        stopColor={entry.color}
                        stopOpacity={1}
                      />
                      <stop
                        offset="100%"
                        stopColor={entry.color}
                        stopOpacity={0.7}
                      />
                    </linearGradient>
                  ))}
                </defs>
                <Pie
                  data={signalData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  outerRadius={90}
                  innerRadius={50}
                  paddingAngle={3}
                  dataKey="value"
                  animationDuration={1500}>
                  {signalData.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={`url(#gradient${index})`}
                      stroke="#fff"
                      strokeWidth={3}
                    />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1f2937',
                    border: 'none',
                    borderRadius: '12px',
                    boxShadow: '0 20px 40px rgba(0,0,0,0.3)',
                    color: '#fff',
                    padding: '12px',
                    fontWeight: 600,
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
            <div className="mt-6 grid grid-cols-3 gap-3">
              {signalData.map((entry, index) => (
                <div
                  key={index}
                  className="flex flex-col items-center p-4 bg-gradient-to-br from-gray-50 to-white rounded-xl hover:shadow-md transition-all border border-gray-100">
                  <div className="flex items-center gap-2 mb-2">
                    <div
                      className="w-3 h-3 rounded-full shadow-lg"
                      style={{backgroundColor: entry.color}}
                    />
                    <span className="text-xs font-bold text-gray-700">
                      {entry.name}
                    </span>
                  </div>
                  <div
                    className="text-3xl font-bold"
                    style={{color: entry.color}}>
                    {entry.value}
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    {((entry.value / totalSignals) * 100).toFixed(0)}%
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
