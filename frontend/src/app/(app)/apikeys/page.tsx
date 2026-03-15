'use client';

import { useState, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { toast } from 'sonner';
import { api, APIKeyItem } from '@/lib/api';
import { Plus, Trash2, Power, PowerOff, Copy, Key } from 'lucide-react';

export default function APIKeysPage() {
  const [keys, setKeys] = useState<APIKeyItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [dialogOpen, setDialogOpen] = useState(false);
  const [serviceMode, setServiceMode] = useState<'all' | 'custom'>('all');
  const [selectedChannels, setSelectedChannels] = useState<Set<string>>(new Set());
  const [submitting, setSubmitting] = useState(false);
  const [latestCreatedKey, setLatestCreatedKey] = useState('');

  const ALL_CHANNELS = [
    { key: 'ocr', label: 'OCR', desc: '文字识别' },
    { key: 'chat', label: 'Chat', desc: '对话模型' },
    { key: 'image', label: 'Image', desc: '图片生成' },
  ];

  const fetchKeys = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getAPIKeys();
      setKeys(data);
      setSelectedIds(new Set());
    } catch {
      toast.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchKeys();
  }, [fetchKeys]);

  const toggleSelect = (id: number) => {
    const newSet = new Set(selectedIds);
    if (newSet.has(id)) newSet.delete(id);
    else newSet.add(id);
    setSelectedIds(newSet);
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === keys.length) setSelectedIds(new Set());
    else setSelectedIds(new Set(keys.map((k) => k.id)));
  };

  const handleCreate = async () => {
    const services = serviceMode === 'all' ? '*' : Array.from(selectedChannels).join(',');
    if (serviceMode === 'custom' && selectedChannels.size === 0) {
      toast.error('请至少选择一个渠道');
      return;
    }
    setSubmitting(true);
    setLatestCreatedKey('');
    try {
      const result = await api.createAPIKey(services);
      setLatestCreatedKey(result.key);
      navigator.clipboard.writeText(result.key);
      toast.success('API Key 创建成功，已复制到剪贴板');
      setDialogOpen(false);
      setServiceMode('all');
      setSelectedChannels(new Set());
      fetchKeys();
    } catch (err) {
      toast.error('创建失败', {
        description: err instanceof Error ? err.message : '',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await api.deleteAPIKey(id);
      toast.success('删除成功');
      fetchKeys();
    } catch {
      toast.error('删除失败');
    }
  };

  const handleToggle = async (id: number, currentEnabled: boolean) => {
    try {
      await api.toggleAPIKey(id);
      toast.success(currentEnabled ? '已禁用' : '已启用');
      fetchKeys();
    } catch {
      toast.error('操作失败');
    }
  };

  const handleBatchDelete = async () => {
    if (selectedIds.size === 0) return;
    try {
      const ids = Array.from(selectedIds);
      await api.batchDeleteAPIKeys(ids);
      toast.success(`成功删除 ${ids.length} 个`);
      fetchKeys();
    } catch {
      toast.error('批量删除失败');
    }
  };

  const handleBatchToggle = async (enable: boolean) => {
    if (selectedIds.size === 0) return;
    try {
      const ids = Array.from(selectedIds);
      await api.batchToggleAPIKeys(ids, enable);
      toast.success(`已${enable ? '启用' : '禁用'} ${ids.length} 个`);
      fetchKeys();
    } catch {
      toast.error('批量操作失败');
    }
  };

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    toast.success('已复制');
  };

  const formatServices = (s: string) => {
    if (s === '*') return '全部';
    return s.split(',').map((v) => v.trim()).join(', ');
  };

  return (
    <div className="space-y-5">
      {/* Action bar */}
      <section className="rounded-3xl border border-border bg-card p-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-2">
            <Checkbox
              checked={selectedIds.size === keys.length && keys.length > 0}
              onCheckedChange={toggleSelectAll}
            />
            <span className="text-xs text-muted-foreground">
              已选 {selectedIds.size} / {keys.length} 个
            </span>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <button
              className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
              onClick={() => setDialogOpen(true)}
            >
              <Plus className="inline h-4 w-4 mr-1" />
              创建 Key
            </button>
            <button
              className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-50"
              disabled={selectedIds.size === 0}
              onClick={() => handleBatchToggle(true)}
            >
              <Power className="inline h-4 w-4 mr-1" />
              批量启用
            </button>
            <button
              className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-50"
              disabled={selectedIds.size === 0}
              onClick={() => handleBatchToggle(false)}
            >
              <PowerOff className="inline h-4 w-4 mr-1" />
              批量禁用
            </button>
            <button
              className="rounded-full border border-destructive/40 px-4 py-2 text-sm font-medium text-destructive transition-colors hover:border-destructive hover:bg-destructive/10 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={selectedIds.size === 0}
              onClick={handleBatchDelete}
            >
              <Trash2 className="inline h-4 w-4 mr-1" />
              批量删除
            </button>
          </div>
        </div>

        {/* API Keys table */}
        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-muted-foreground">
                <th className="px-3 py-2 w-10"></th>
                <th className="px-3 py-2">ID</th>
                <th className="px-3 py-2">Key</th>
                <th className="px-3 py-2">服务类型</th>
                <th className="px-3 py-2">创建时间</th>
                <th className="px-3 py-2">状态</th>
                <th className="px-3 py-2">操作</th>
              </tr>
            </thead>
            <tbody>
              {keys.map((item) => (
                <tr key={item.id} className="border-b border-border/50">
                  <td className="px-3 py-2">
                    <Checkbox
                      checked={selectedIds.has(item.id)}
                      onCheckedChange={() => toggleSelect(item.id)}
                    />
                  </td>
                  <td className="px-3 py-2 text-foreground">{item.id}</td>
                  <td className="px-3 py-2">
                    <div className="flex items-center gap-2">
                      <code className="max-w-[200px] truncate text-xs text-foreground">{item.key}</code>
                      <button
                        className="rounded-lg border border-border px-2 py-0.5 text-[11px] text-muted-foreground transition-colors hover:border-primary hover:text-primary"
                        onClick={() => copyKey(item.key)}
                      >
                        复制
                      </button>
                    </div>
                  </td>
                  <td className="px-3 py-2">
                    <span className="inline-flex items-center gap-1 rounded-full border border-border px-2 py-0.5 text-[11px] text-foreground">
                      <Key className="h-3 w-3" />
                      {formatServices(item.services)}
                    </span>
                  </td>
                  <td className="px-3 py-2 text-muted-foreground whitespace-nowrap">
                    {new Date(item.created_at).toLocaleString('zh-CN')}
                  </td>
                  <td className="px-3 py-2">
                    <span className={item.enabled ? 'text-emerald-600' : 'text-destructive'}>
                      {item.enabled ? '启用' : '禁用'}
                    </span>
                  </td>
                  <td className="px-3 py-2">
                    <div className="flex items-center gap-1">
                      <button
                        className="rounded-xl border border-border px-3 py-1 text-xs text-foreground transition-colors hover:bg-accent"
                        onClick={() => handleToggle(item.id, item.enabled)}
                      >
                        {item.enabled ? '禁用' : '启用'}
                      </button>
                      <button
                        className="rounded-xl border border-destructive/30 px-3 py-1 text-xs text-destructive transition-colors hover:bg-destructive/10"
                        onClick={() => handleDelete(item.id)}
                      >
                        删除
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {loading && (
            <div className="py-8 text-center text-sm text-muted-foreground">加载中...</div>
          )}
          {!loading && keys.length === 0 && (
            <div className="py-8 text-center text-sm text-muted-foreground">暂无 API Key</div>
          )}
        </div>
      </section>

      {/* Create Dialog */}
      {dialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4" onClick={() => !submitting && setDialogOpen(false)}>
          <div className="w-full max-w-lg rounded-3xl border border-border bg-card p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-foreground">创建 API Key</p>
              <button className="text-xs text-muted-foreground hover:text-foreground" onClick={() => setDialogOpen(false)}>关闭</button>
            </div>
            <p className="mt-2 text-xs text-muted-foreground">选择该 Key 可访问的服务类型</p>
            <div className="mt-4 space-y-4">
              <div className="space-y-2">
                <label className="block text-xs text-muted-foreground">服务范围</label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    className={`rounded-2xl border px-4 py-3 text-left text-sm transition-colors ${
                      serviceMode === 'all'
                        ? 'border-primary bg-primary/5 text-primary ring-1 ring-primary/30'
                        : 'border-border text-foreground hover:border-primary/50'
                    }`}
                    onClick={() => setServiceMode('all')}
                  >
                    <p className="font-medium">全部渠道</p>
                    <p className="mt-0.5 text-[11px] text-muted-foreground">OCR / Chat / Image</p>
                  </button>
                  <button
                    type="button"
                    className={`rounded-2xl border px-4 py-3 text-left text-sm transition-colors ${
                      serviceMode === 'custom'
                        ? 'border-primary bg-primary/5 text-primary ring-1 ring-primary/30'
                        : 'border-border text-foreground hover:border-primary/50'
                    }`}
                    onClick={() => setServiceMode('custom')}
                  >
                    <p className="font-medium">指定渠道</p>
                    <p className="mt-0.5 text-[11px] text-muted-foreground">按需选择可用服务</p>
                  </button>
                </div>
              </div>
              {serviceMode === 'custom' && (
                <div className="space-y-2">
                  <label className="block text-xs text-muted-foreground">选择渠道</label>
                  <div className="grid grid-cols-2 gap-2">
                    {ALL_CHANNELS.map((ch) => (
                      <button
                        key={ch.key}
                        type="button"
                        className={`rounded-2xl border px-4 py-3 text-left text-sm transition-colors ${
                          selectedChannels.has(ch.key)
                            ? 'border-primary bg-primary/5 text-primary ring-1 ring-primary/30'
                            : 'border-border text-foreground hover:border-primary/50'
                        }`}
                        onClick={() => {
                          const next = new Set(selectedChannels);
                          if (next.has(ch.key)) next.delete(ch.key);
                          else next.add(ch.key);
                          setSelectedChannels(next);
                        }}
                      >
                        <div className="flex items-center gap-2">
                          <Checkbox checked={selectedChannels.has(ch.key)} />
                          <p className="font-medium">{ch.label}</p>
                        </div>
                        <p className="mt-0.5 text-[11px] text-muted-foreground">{ch.desc}</p>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
            <div className="mt-6 flex items-center justify-end gap-2">
              <button
                className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-accent"
                onClick={() => setDialogOpen(false)}
                disabled={submitting}
              >
                取消
              </button>
              <button
                className="rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
                onClick={handleCreate}
                disabled={submitting || (serviceMode === 'custom' && selectedChannels.size === 0)}
              >
                {submitting ? '创建中...' : '创建 Key'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
