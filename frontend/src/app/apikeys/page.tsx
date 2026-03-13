'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { toast } from 'sonner';
import { api, APIKeyItem } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { Plus, Trash2, Power, PowerOff, Copy, LogOut, Key } from 'lucide-react';

export default function APIKeysPage() {
  const router = useRouter();
  const { isAuthenticated, logout } = useAuth();

  const [keys, setKeys] = useState<APIKeyItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

  const [dialogOpen, setDialogOpen] = useState(false);
  const [services, setServices] = useState('*');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

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
    if (isAuthenticated) fetchKeys();
  }, [isAuthenticated, fetchKeys]);

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
    setSubmitting(true);
    try {
      const result = await api.createAPIKey(services);
      toast.success('API Key 创建成功');
      // 复制到剪贴板
      navigator.clipboard.writeText(result.key);
      toast.info('Key 已复制到剪贴板');
      setDialogOpen(false);
      setServices('*');
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

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-950 p-6">
      <div className="mx-auto max-w-6xl space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">API Key 管理</h1>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => router.push('/tokens')}>
              Token 管理
            </Button>
            <Button variant="outline" onClick={() => router.push('/logs')}>
              请求日志
            </Button>
            <Button variant="outline" onClick={handleLogout}>
              <LogOut className="mr-2 h-4 w-4" />
              退出
            </Button>
          </div>
        </div>

        {/* Toolbar */}
        <div className="flex flex-wrap items-center gap-2">
          <Button onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            创建 Key
          </Button>
          <Button
            variant="outline"
            onClick={() => handleBatchToggle(true)}
            disabled={selectedIds.size === 0}
          >
            <Power className="mr-2 h-4 w-4" />
            批量启用
          </Button>
          <Button
            variant="outline"
            onClick={() => handleBatchToggle(false)}
            disabled={selectedIds.size === 0}
          >
            <PowerOff className="mr-2 h-4 w-4" />
            批量禁用
          </Button>
          <Button
            variant="destructive"
            onClick={handleBatchDelete}
            disabled={selectedIds.size === 0}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            批量删除
          </Button>
        </div>

        {/* Table */}
        {loading ? (
          <div className="text-center py-8 text-neutral-500">加载中...</div>
        ) : keys.length === 0 ? (
          <div className="text-center py-8 text-neutral-500">暂无 API Key</div>
        ) : (
          <div className="rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <Checkbox
                      checked={selectedIds.size === keys.length && keys.length > 0}
                      onCheckedChange={toggleSelectAll}
                    />
                  </TableHead>
                  <TableHead>ID</TableHead>
                  <TableHead>Key</TableHead>
                  <TableHead>服务类型</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <Checkbox
                        checked={selectedIds.has(item.id)}
                        onCheckedChange={() => toggleSelect(item.id)}
                      />
                    </TableCell>
                    <TableCell>{item.id}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <code className="max-w-[200px] truncate text-sm">
                          {item.key}
                        </code>
                        <Button
                          variant="ghost"
                          size="icon-xs"
                          onClick={() => copyKey(item.key)}
                          title="复制"
                        >
                          <Copy className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-2 py-1 text-xs font-medium text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
                        <Key className="h-3 w-3" />
                        {formatServices(item.services)}
                      </span>
                    </TableCell>
                    <TableCell>
                      {new Date(item.created_at).toLocaleString('zh-CN')}
                    </TableCell>
                    <TableCell>
                      <span
                        className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
                          item.enabled
                            ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                            : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                        }`}
                      >
                        {item.enabled ? '启用' : '禁用'}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon-xs"
                          onClick={() => handleToggle(item.id, item.enabled)}
                        >
                          {item.enabled ? (
                            <PowerOff className="h-3.5 w-3.5" />
                          ) : (
                            <Power className="h-3.5 w-3.5" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon-xs"
                          onClick={() => handleDelete(item.id)}
                          className="text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>

      {/* Create Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>创建 API Key</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">服务类型</label>
              <Input
                placeholder="* 表示全部，或 ocr,audio,chat"
                value={services}
                onChange={(e) => setServices(e.target.value)}
              />
              <p className="text-xs text-neutral-500 mt-1">
                * = 全部服务 | 多个服务用逗号分隔，如: ocr,audio
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleCreate} disabled={submitting}>
              {submitting ? '创建中...' : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
