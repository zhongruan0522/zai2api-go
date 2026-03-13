'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { toast } from 'sonner';
import { api, LogItem, LogStats } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { LogOut, RefreshCw } from 'lucide-react';

export default function LogsPage() {
  const router = useRouter();
  const { isAuthenticated, logout } = useAuth();

  const [logs, setLogs] = useState<LogItem[]>([]);
  const [stats, setStats] = useState<LogStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [channel, setChannel] = useState('');

  useEffect(() => {
    if (!isAuthenticated) router.push('/login');
  }, [isAuthenticated, router]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [logsRes, statsRes] = await Promise.all([
        api.getLogs(channel || undefined),
        api.getLogStats(),
      ]);
      setLogs(logsRes.data);
      setStats(statsRes);
    } catch {
      toast.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  }, [channel]);

  useEffect(() => {
    if (isAuthenticated) fetchData();
  }, [isAuthenticated, fetchData]);

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  const formatTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-950 p-6">
      <div className="mx-auto max-w-6xl space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">请求日志</h1>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => router.push('/tokens')}>
              Token 管理
            </Button>
            <Button variant="outline" onClick={() => router.push('/apikeys')}>
              API Key
            </Button>
            <Button variant="outline" onClick={handleLogout}>
              <LogOut className="mr-2 h-4 w-4" />
              退出
            </Button>
          </div>
        </div>

        {/* Stats */}
        {stats && (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 lg:grid-cols-7">
            <StatCard label="总请求" value={stats.total} />
            <StatCard label="成功" value={stats.success} color="green" />
            <StatCard label="失败" value={stats.failed} color="red" />
            <StatCard label="今日" value={stats.today} color="blue" />
            <StatCard label="OCR" value={stats.ocr} />
            <StatCard label="Audio" value={stats.audio} />
            <StatCard label="Chat" value={stats.chat} />
          </div>
        )}

        {/* Filter */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-neutral-500">渠道筛选：</span>
          {['', 'ocr', 'audio', 'chat'].map((ch) => (
            <Button
              key={ch}
              variant={channel === ch ? 'default' : 'outline'}
              size="sm"
              onClick={() => setChannel(ch)}
            >
              {ch || '全部'}
            </Button>
          ))}
          <Button variant="outline" size="sm" onClick={fetchData}>
            <RefreshCw className="h-3.5 w-3.5" />
          </Button>
        </div>

        {/* Table */}
        {loading ? (
          <div className="text-center py-8 text-neutral-500">加载中...</div>
        ) : logs.length === 0 ? (
          <div className="text-center py-8 text-neutral-500">暂无日志</div>
        ) : (
          <div className="rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>请求 ID</TableHead>
                  <TableHead>时间</TableHead>
                  <TableHead>渠道</TableHead>
                  <TableHead>源 IP</TableHead>
                  <TableHead>API Key ID</TableHead>
                  <TableHead>Token ID</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>错误信息</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {logs.map((log) => (
                  <TableRow key={log.id}>
                    <TableCell className="font-mono text-xs max-w-[120px] truncate">
                      {log.request_id.slice(0, 8)}...
                    </TableCell>
                    <TableCell className="text-sm whitespace-nowrap">
                      {formatTime(log.created_at)}
                    </TableCell>
                    <TableCell>
                      <span className="inline-flex items-center rounded-full bg-neutral-100 px-2 py-1 text-xs font-medium dark:bg-neutral-800">
                        {log.channel}
                      </span>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {log.source_ip}
                    </TableCell>
                    <TableCell>{log.api_key_id}</TableCell>
                    <TableCell>{log.token_id}</TableCell>
                    <TableCell>
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
                          log.success
                            ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                            : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                        }`}
                      >
                        {log.success ? '成功' : '失败'}
                      </span>
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate text-sm text-neutral-500">
                      {log.error_msg || '-'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color?: string }) {
  const colorMap: Record<string, string> = {
    green: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    red: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
    blue: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  };
  return (
    <div className="rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-4 text-center">
      <div className={`text-2xl font-bold ${colorMap[color || ''] || ''}`}>{value}</div>
      <div className="text-sm text-neutral-500">{label}</div>
    </div>
  );
}
