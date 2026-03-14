'use client';

import { Activity } from 'lucide-react';

export default function MonitorPage() {
  return (
    <div className="space-y-5">
      {/* Placeholder header */}
      <section className="rounded-3xl border border-border bg-card p-6">
        <div className="flex items-center gap-3 mb-4">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded-2xl border border-border bg-primary/15 text-primary border-primary/30">
            <Activity className="h-5 w-5" />
          </span>
          <div>
            <p className="text-base font-semibold text-foreground">服务监控</p>
            <p className="text-xs text-muted-foreground">实时监控服务运行状态</p>
          </div>
        </div>

        <div className="mt-6 rounded-2xl border border-dashed border-border/60 bg-secondary/30 p-8 text-center">
          <Activity className="mx-auto h-12 w-12 text-muted-foreground/30" />
          <p className="mt-4 text-sm font-medium text-muted-foreground">
            服务监控功能开发中
          </p>
          <p className="mt-2 text-xs text-muted-foreground/60">
            该功能将展示各渠道（Audio / OCR / Chat）的可用性、响应时间和成功率等监控数据
          </p>
          <span className="mt-4 inline-flex items-center rounded-full border border-border bg-card px-3 py-1 text-[11px] text-muted-foreground">
            演示页面 · 功能待实现
          </span>
        </div>
      </section>

      {/* Channel status placeholders */}
      <section className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {[
          { name: 'Audio 渠道', color: 'text-blue-600', bgColor: 'bg-blue-100' },
          { name: 'OCR 渠道', color: 'text-violet-600', bgColor: 'bg-violet-100' },
          { name: 'Chat 渠道', color: 'text-teal-600', bgColor: 'bg-teal-100' },
        ].map((channel) => (
          <div key={channel.name} className="rounded-3xl border border-border bg-card p-5">
            <p className="text-sm font-medium text-foreground">{channel.name}</p>
            <div className="mt-4 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs text-muted-foreground">服务状态</span>
                <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium ${channel.bgColor} ${channel.color}`}>
                  待接入
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs text-muted-foreground">可用率</span>
                <span className="text-xs text-muted-foreground">--</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs text-muted-foreground">平均响应</span>
                <span className="text-xs text-muted-foreground">--</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs text-muted-foreground">可用 Token</span>
                <span className="text-xs text-muted-foreground">--</span>
              </div>
            </div>
          </div>
        ))}
      </section>
    </div>
  );
}
