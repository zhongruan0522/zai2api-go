'use client';

import { useState, useEffect, useCallback } from 'react';
import { toast } from 'sonner';
import { api, LogItem, LogStats } from '@/lib/api';
import { RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react';

const PAGE_SIZE_OPTIONS = [10, 50, 100];

export default function LogsPage() {
  const [logs, setLogs] = useState<LogItem[]>([]);
  const [stats, setStats] = useState<LogStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [channel, setChannel] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [total, setTotal] = useState(0);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [logsRes, statsRes] = await Promise.all([
        api.getLogs(channel || undefined, page, pageSize),
        api.getLogStats(),
      ]);
      setLogs(logsRes.data);
      setTotal(logsRes.total);
      setStats(statsRes);
      if (page > totalPages && totalPages > 0) {
        setPage(totalPages);
      }
    } catch {
      toast.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  }, [channel, page, pageSize, totalPages]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleChannelChange = (ch: string) => {
    setChannel(ch);
    setPage(1);
  };

  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setPage(1);
  };

  const formatTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  const goPage = (p: number) => {
    if (p < 1 || p > totalPages) return;
    setPage(p);
  };

  const getPageNumbers = () => {
    const pages: number[] = [];
    const maxVisible = 5;
    let start = Math.max(1, page - Math.floor(maxVisible / 2));
    let end = Math.min(totalPages, start + maxVisible - 1);
    if (end - start + 1 < maxVisible) {
      start = Math.max(1, end - maxVisible + 1);
    }
    for (let i = start; i <= end; i++) {
      pages.push(i);
    }
    return pages;
  };

  return (
    <div className="space-y-5">
      {stats && (
        <section className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
          <StatCard label="总请求" value={stats.total} iconBg="bg-sky-100" />
          <StatCard label="成功" value={stats.success} iconBg="bg-emerald-100" />
          <StatCard label="失败" value={stats.failed} iconBg="bg-red-100" />
          <StatCard label="今日" value={stats.today} iconBg="bg-amber-100" />
          <StatCard label="OCR" value={stats.ocr} iconBg="bg-violet-100" />
          <StatCard label="Audio" value={stats.audio} iconBg="bg-blue-100" />
          <StatCard label="Chat" value={stats.chat} iconBg="bg-teal-100" />
          <StatCard label="Image" value={stats.image} iconBg="bg-pink-100" />
        </section>
      )}

      <section className="rounded-3xl border border-border bg-card p-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <p className="text-base font-semibold text-foreground">请求日志</p>
          <div className="flex flex-wrap items-center gap-2">
            {['', 'ocr', 'audio', 'chat', 'image'].map((ch) => (
              <button
                key={ch}
                className={`rounded-full border px-4 py-2 text-xs font-medium transition-all ${
                  channel === ch
                    ? 'bg-accent text-foreground border-primary/50 font-semibold'
                    : 'bg-transparent text-muted-foreground hover:bg-accent/50 hover:text-foreground border-border'
                }`}
                onClick={() => handleChannelChange(ch)}
              >
                {ch || '全部'}
              </button>
            ))}
            <button
              className="rounded-full border border-border px-4 py-2 text-xs font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
              onClick={fetchData}
            >
              <RefreshCw className="inline h-3.5 w-3.5 mr-1" />
              刷新
            </button>
          </div>
        </div>

        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-muted-foreground">
                <th className="px-3 py-2">请求 ID</th>
                <th className="px-3 py-2">时间</th>
                <th className="px-3 py-2">渠道</th>
                <th className="px-3 py-2">源 IP</th>
                <th className="px-3 py-2">API Key</th>
                <th className="px-3 py-2">Token</th>
                <th className="px-3 py-2">状态</th>
                <th className="px-3 py-2">错误信息</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log) => (
                <tr key={log.id} className="border-b border-border/50">
                  <td className="px-3 py-2 font-mono text-xs max-w-[120px] truncate text-foreground">
                    {log.request_id.slice(0, 8)}...
                  </td>
                  <td className="px-3 py-2 text-xs whitespace-nowrap text-muted-foreground">
                    {formatTime(log.created_at)}
                  </td>
                  <td className="px-3 py-2">
                    <span className="inline-flex items-center rounded-full border border-border px-2 py-0.5 text-[11px] font-medium text-foreground">
                      {log.channel}
                    </span>
                  </td>
                  <td className="px-3 py-2 font-mono text-xs text-foreground">{log.source_ip}</td>
                  <td className="px-3 py-2 text-muted-foreground">{log.api_key_id}</td>
                  <td className="px-3 py-2 text-muted-foreground">{log.token_id}</td>
                  <td className="px-3 py-2">
                    <span className={log.success ? 'text-emerald-600' : 'text-destructive'}>
                      {log.success ? '成功' : '失败'}
                    </span>
                  </td>
                  <td className="px-3 py-2 max-w-[200px] truncate text-xs text-muted-foreground">
                    {log.error_msg || '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {loading && (
            <div className="py-8 text-center text-sm text-muted-foreground">加载中...</div>
          )}
          {!loading && logs.length === 0 && (
            <div className="py-8 text-center text-sm text-muted-foreground">暂无日志</div>
          )}
        </div>

        {!loading && total > 0 && (
          <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-4">
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">每页</span>
              {PAGE_SIZE_OPTIONS.map((size) => (
                <button
                  key={size}
                  className={`rounded-full border px-3 py-1 text-xs font-medium transition-all ${
                    pageSize === size
                      ? 'bg-accent text-foreground border-primary/50'
                      : 'bg-transparent text-muted-foreground hover:text-foreground border-border'
                  }`}
                  onClick={() => handlePageSizeChange(size)}
                >
                  {size}
                </button>
              ))}
              <span className="text-xs text-muted-foreground">
                共 {total} 条，第 {page}/{totalPages} 页
              </span>
            </div>
            <div className="flex items-center gap-1">
              <button
                className="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-40"
                disabled={page <= 1}
                onClick={() => goPage(1)}
              >
                首页
              </button>
              <button
                className="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-40"
                disabled={page <= 1}
                onClick={() => goPage(page - 1)}
              >
                <ChevronLeft className="h-4 w-4" />
              </button>
              {getPageNumbers().map((p) => (
                <button
                  key={p}
                  className={`inline-flex h-8 w-8 items-center justify-center rounded-full border text-xs font-medium transition-all ${
                    p === page
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'border-border text-muted-foreground hover:border-primary hover:text-foreground'
                  }`}
                  onClick={() => goPage(p)}
                >
                  {p}
                </button>
              ))}
              <button
                className="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-40"
                disabled={page >= totalPages}
                onClick={() => goPage(page + 1)}
              >
                <ChevronRight className="h-4 w-4" />
              </button>
              <button
                className="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted-foreground transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-40"
                disabled={page >= totalPages}
                onClick={() => goPage(totalPages)}
              >
                末页
              </button>
            </div>
          </div>
        )}
      </section>
    </div>
  );
}

function StatCard({ label, value, iconBg }: { label: string; value: number; iconBg: string }) {
  return (
    <div className="rounded-3xl border border-border bg-card p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <p className="text-xs uppercase tracking-[0.3em] text-muted-foreground">{label}</p>
          <p className="mt-2 text-2xl font-semibold text-foreground tabular-nums">
            {value.toLocaleString()}
          </p>
        </div>
        <div className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full ${iconBg}`}>
        </div>
      </div>
    </div>
  );
}
