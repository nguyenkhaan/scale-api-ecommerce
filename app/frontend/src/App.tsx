import React, { useState, useEffect } from 'react';
import { fetchHealth, fetchPing, ApiResponse, HealthData } from './api/client';

export const App: React.FC = () => {
  const [health, setHealth] = useState<ApiResponse<HealthData> | null>(null);
  const [pingResult, setPingResult] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [connected, setConnected] = useState<boolean | null>(null);

  const checkConnection = async () => {
    setLoading(true);
    setError(null);
    try {
      const [h, p] = await Promise.all([fetchHealth(), fetchPing()]);
      setHealth(h);
      if (p.data) {
        setPingResult(p.data.message);
      }
      setConnected(h.success);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Không thể kết nối đến Backend API');
      setConnected(false);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkConnection();
  }, []);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col justify-between p-4 md:p-8 font-sans">
      <header className="max-w-5xl mx-auto w-full flex items-center justify-between py-4 border-b border-slate-800">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-500 to-rose-500 flex items-center justify-center font-bold text-xl shadow-lg shadow-indigo-500/20">
            ⚡
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent">
              SaleScale
            </h1>
            <p className="text-xs text-slate-400">Hệ thống Flash Sale Phân Tán</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <span
            className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium border ${
              connected === true
                ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                : connected === false
                ? 'bg-rose-500/10 text-rose-400 border-rose-500/30'
                : 'bg-amber-500/10 text-amber-400 border-amber-500/30'
            }`}
          >
            <span
              className={`w-2 h-2 rounded-full ${
                connected === true
                  ? 'bg-emerald-400 animate-pulse'
                  : connected === false
                  ? 'bg-rose-400'
                  : 'bg-amber-400'
              }`}
            />
            {connected === true
              ? 'Backend Online'
              : connected === false
              ? 'Backend Offline'
              : 'Đang kết nối...'}
          </span>
        </div>
      </header>

      <main className="max-w-5xl mx-auto w-full my-8 space-y-6">
        {/* Hero Banner */}
        <div className="relative overflow-hidden rounded-2xl bg-gradient-to-b from-slate-900 to-slate-950 border border-slate-800 p-8 shadow-2xl">
          <div className="absolute top-0 right-0 -mt-10 -mr-10 w-72 h-72 rounded-full bg-indigo-500/10 blur-3xl pointer-events-none" />
          <div className="relative z-10 space-y-4">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-xs font-semibold">
              Frontend Scaffold • React 19 + Vite + Tailwind CSS
            </div>
            <h2 className="text-3xl md:text-4xl font-extrabold tracking-tight text-white">
              Sẵn sàng cho bài toán tải cao & Flash Sale
            </h2>
            <p className="text-slate-400 max-w-2xl text-sm md:text-base leading-relaxed">
              Kiến trúc kết nối trực tiếp đến backend <span className="text-indigo-300 font-mono">Go + Gin</span> qua tiền tố chuẩn <span className="text-indigo-300 font-mono">/api</span>. Nền tảng đảm bảo giữ chỗ 5 phút, quota nguyên tử, và không bao giờ oversell.
            </p>

            <div className="flex flex-wrap items-center gap-3 pt-2">
              <button
                onClick={checkConnection}
                disabled={loading}
                className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-medium text-sm rounded-lg shadow-md shadow-indigo-600/20 transition-all cursor-pointer flex items-center gap-2"
              >
                {loading ? 'Đang kiểm tra...' : '🔄 Kiểm tra kết nối API'}
              </button>
              <a
                href="/api/health"
                target="_blank"
                rel="noreferrer"
                className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 font-medium text-sm rounded-lg border border-slate-700 transition-all"
              >
                Mở /api/health ↗
              </a>
            </div>
          </div>
        </div>

        {/* Status & Tech Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Card: Backend API Status */}
          <div className="bg-slate-900/60 border border-slate-800/80 rounded-2xl p-6 backdrop-blur space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-base font-semibold text-white flex items-center gap-2">
                <span>📡</span> Trạng thái Backend API
              </h3>
              <span className="text-xs font-mono text-slate-400">Prefix: /api</span>
            </div>

            {error && (
              <div className="p-3.5 bg-rose-500/10 border border-rose-500/20 rounded-xl text-rose-400 text-xs">
                {error}
                <div className="mt-1 text-slate-400">
                  (Hãy đảm bảo backend đã chạy bằng lệnh <code className="text-white font-mono">npm run dev</code> ở <code className="text-white font-mono">app/backend</code>)
                </div>
              </div>
            )}

            {health && (
              <div className="space-y-3 font-mono text-xs">
                <div className="grid grid-cols-2 gap-2 bg-slate-950/70 p-3 rounded-xl border border-slate-800/50">
                  <div className="text-slate-400">Dịch vụ:</div>
                  <div className="text-slate-200 font-semibold">{health.data?.service}</div>

                  <div className="text-slate-400">Môi trường:</div>
                  <div className="text-indigo-400 font-semibold uppercase">{health.data?.environment}</div>

                  <div className="text-slate-400">Uptime:</div>
                  <div className="text-emerald-400 font-semibold">{health.data?.uptime}</div>

                  <div className="text-slate-400">Ping:</div>
                  <div className="text-slate-200 font-semibold">{pingResult || 'N/A'}</div>
                </div>

                <div className="text-slate-500 text-[11px] truncate">
                  Request ID: <span className="text-slate-400">{health.meta.request_id}</span>
                </div>
              </div>
            )}
          </div>

          {/* Card: Tech Stack Info */}
          <div className="bg-slate-900/60 border border-slate-800/80 rounded-2xl p-6 backdrop-blur space-y-4">
            <h3 className="text-base font-semibold text-white flex items-center gap-2">
              <span>🛠️</span> Công nghệ đã thiết lập
            </h3>

            <div className="space-y-2.5 text-xs">
              <div className="flex items-center justify-between p-2.5 bg-slate-950/50 rounded-lg border border-slate-800/40">
                <span className="text-slate-300 font-medium">Frontend Framework</span>
                <span className="font-mono text-indigo-400">React 19 + Vite</span>
              </div>
              <div className="flex items-center justify-between p-2.5 bg-slate-950/50 rounded-lg border border-slate-800/40">
                <span className="text-slate-300 font-medium">CSS Utility</span>
                <span className="font-mono text-sky-400">Tailwind CSS v4</span>
              </div>
              <div className="flex items-center justify-between p-2.5 bg-slate-950/50 rounded-lg border border-slate-800/40">
                <span className="text-slate-300 font-medium">Backend API</span>
                <span className="font-mono text-emerald-400">Go 1.24 + Gin</span>
              </div>
              <div className="flex items-center justify-between p-2.5 bg-slate-950/50 rounded-lg border border-slate-800/40">
                <span className="text-slate-300 font-medium">Cơ sở dữ liệu chính</span>
                <span className="font-mono text-amber-400">PostgreSQL 17</span>
              </div>
            </div>
          </div>
        </div>
      </main>

      <footer className="max-w-5xl mx-auto w-full pt-4 border-t border-slate-800/60 text-center text-xs text-slate-500">
        SaleScale Project • UIT 2026 • Phân tán Flash Sale Platform
      </footer>
    </div>
  );
};

