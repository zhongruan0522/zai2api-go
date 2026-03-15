'use client';

import { useCallback, useEffect, useState } from 'react';
import { api, ChannelAvailability, ChannelTrends, MonitorChannels, MonitorSummary } from '@/lib/api';
import { Activity, RefreshCw } from 'lucide-react';
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

const CHANNELS = [
  { key: 'ocr' as const, label: 'OCR', color: '#8b5cf6' },
  { key: 'chat' as const, label: 'Chat', color: '#14b8a6' },
  { key: 'image' as const, label: 'Image', color: '#f97316' },
];

function availColor(avail: number | undefined): string {
  if (avail === undefined || avail === null) return 'text-muted-foreground';
  if (avail >= 99) return 'text-emerald-600';
  if (avail >= 95) return 'text-amber-600';
  return 'text-destructive';
}

function availLabel(avail: number | undefined): string {
  if (avail === undefined || avail === null) return '--';
  return `${avail}%`;
}

function WindowStats({
  title,
  data,
  loading,
}: {
  title: string;
  data?: ChannelAvailability;
  loading: boolean;
}) {
  return (
    <div className="rounded-2xl border border-border/70 bg-background/60 p-4">
      <div className="flex items-center justify-between gap-3">
        <span className="text-xs text-muted-foreground">{title}</span>
        <span className={`text-sm font-semibold tabular-nums ${availColor(data?.availability)}`}>
          {loading ? '--' : availLabel(data?.availability)}
        </span>
      </div>
      <div className="mt-3 grid grid-cols-3 gap-2 text-xs">
        <div>
          <p className="text-muted-foreground">请求</p>
          <p className="mt-1 font-medium text-foreground tabular-nums">{loading ? '--' : (data?.total ?? 0).toLocaleString()}</p>
        </div>
        <div>
          <p className="text-muted-foreground">成功</p>
          <p className="mt-1 font-medium text-emerald-600 tabular-nums">{loading ? '--' : (data?.success ?? 0).toLocaleString()}</p>
        </div>
        <div>
          <p className="text-muted-foreground">失败</p>
          <p className="mt-1 font-medium text-destructive tabular-nums">{loading ? '--' : (data?.failed ?? 0).toLocaleString()}</p>
        </div>
      </div>
    </div>
  );
}

function AvailabilityCard({
  channel,
  summary,
  loading,
}: {
  channel: typeof CHANNELS[number];
  summary?: MonitorSummary;
  loading: boolean;
}) {
  const recentHour = summary?.recent_hour[channel.key];
  const today = summary?.today[channel.key];

  return (
    <div className="rounded-3xl border border-border bg-card p-5">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm font-semibold text-foreground">{channel.label} 渠道</p>
        <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: channel.color }} />
      </div>
      <div className="mt-4 space-y-3">
        <WindowStats title="最近 1 小时（上一整点区间）" data={recentHour} loading={loading} />
        <WindowStats title="今日累计（00:00 - 当前小时）" data={today} loading={loading} />
      </div>
    </div>
  );
}

function TodayTrendChart({ data, loading }: { data?: ChannelTrends; loading: boolean }) {
  if (!data || loading) {
    return (
      <div className="rounded-3xl border border-border bg-card p-6">
        <p className="text-sm font-semibold text-foreground">今日按小时可用率</p>
        <div className="mt-8 flex h-72 items-center justify-center text-sm text-muted-foreground">加载中...</div>
      </div>
    );
  }

  const merged = data.ocr.map((item, i) => ({
    label: item.label,
    OCR: item.availability,
    Chat: data.chat[i]?.availability ?? null,
    Image: data.image[i]?.availability ?? null,
    OCRTotal: item.total,
    ChatTotal: data.chat[i]?.total ?? 0,
    ImageTotal: data.image[i]?.total ?? 0,
  }));

  const hasData = merged.some((item) => item.OCRTotal > 0 || item.ChatTotal > 0 || item.ImageTotal > 0);

  return (
    <div className="rounded-3xl border border-border bg-card p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-sm font-semibold text-foreground">今日按小时可用率</p>
          <p className="mt-1 text-xs text-muted-foreground">按自然日展示，小时标签与整点对齐</p>
        </div>
      </div>
      <div className="mt-4 h-72">
        {hasData ? (
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={merged} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis dataKey="label" tick={{ fontSize: 11 }} stroke="var(--muted-foreground)" />
              <YAxis domain={[0, 100]} tick={{ fontSize: 11 }} stroke="var(--muted-foreground)" tickFormatter={(value: number) => `${value}%`} />
              <Tooltip
                contentStyle={{
                  borderRadius: 12,
                  border: '1px solid var(--border)',
                  backgroundColor: 'var(--card)',
                  fontSize: 12,
                }}
                formatter={(value) => (value != null ? `${value}%` : '--')}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Line type="monotone" dataKey="OCR" stroke="#8b5cf6" strokeWidth={2} dot={false} connectNulls />
              <Line type="monotone" dataKey="Chat" stroke="#14b8a6" strokeWidth={2} dot={false} connectNulls />
              <Line type="monotone" dataKey="Image" stroke="#f97316" strokeWidth={2} dot={false} connectNulls />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">今日暂无数据</div>
        )}
      </div>
    </div>
  );
}

export default function MonitorPage() {
  const [summary, setSummary] = useState<MonitorSummary | null>(null);
  const [hourly, setHourly] = useState<ChannelTrends | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const [summaryData, hourlyData] = await Promise.all([
        api.getMonitorSummary(),
        api.getMonitorHourly(),
      ]);
      setSummary(summaryData);
      setHourly(hourlyData);
    } catch {
      // keep stale data
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  return (
    <div className="space-y-5">
      <section className="rounded-3xl border border-border bg-card p-6">
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <span className="inline-flex h-9 w-9 items-center justify-center rounded-2xl border border-primary/30 bg-primary/15 text-primary">
              <Activity className="h-5 w-5" />
            </span>
            <div>
              <p className="text-base font-semibold text-foreground">服务监控</p>
              <p className="text-xs text-muted-foreground">最近 1 小时按整点统计，今日按自然日小时展示</p>
            </div>
          </div>
          <button
            className="rounded-full border border-border px-4 py-2 text-xs font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
            onClick={fetchAll}
          >
            <RefreshCw className="mr-1 inline h-3.5 w-3.5" />
            刷新
          </button>
        </div>
      </section>

      <section className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {CHANNELS.map((channel) => (
          <AvailabilityCard key={channel.key} channel={channel} summary={summary ?? undefined} loading={loading} />
        ))}
      </section>

      <TodayTrendChart data={hourly ?? undefined} loading={loading} />
    </div>
  );
}
