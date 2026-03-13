'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { Toaster, toast } from 'sonner';
import { api, TokenItem } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { Plus, Trash2, Power, PowerOff, LogOut, Key, FileText } from 'lucide-react';

type TokenType = 'audio' | 'ocr' | 'chat';

export default function TokensPage() {
  const router = useRouter();
  const { isAuthenticated, logout } = useAuth();

  const [activeTab, setActiveTab] = useState<TokenType>('audio');
  const [tokens, setTokens] = useState<TokenItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

  const [dialogOpen, setDialogOpen] = useState(false);
  const [newTokens, setNewTokens] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // 检查认证状态
  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  // 获取 tokens
  const fetchTokens = useCallback(async () => {
    setLoading(true);
    try {
      let data: TokenItem[];
      switch (activeTab) {
        case 'audio':
          data = await api.getAudioTokens();
          break;
        case 'ocr':
          data = await api.getOCRTokens();
          break;
        case 'chat':
          data = await api.getChatTokens();
          break;
      }
      setTokens(data);
      setSelectedIds(new Set());
    } catch {
      toast.error('获取数据失败');
    } finally {
      setLoading(false);
    }
  }, [activeTab]);

  useEffect(() => {
    if (isAuthenticated) {
      fetchTokens();
    }
  }, [activeTab, isAuthenticated, fetchTokens]);

  // 选择/取消选择
  const toggleSelect = (id: number) => {
    const newSet = new Set(selectedIds);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setSelectedIds(newSet);
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === tokens.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(tokens.map((t) => t.id)));
    }
  };

  // 导入 tokens
  const handleImport = async () => {
    const tokensList = newTokens
      .split('\n')
      .map((t) => t.trim())
      .filter((t) => t);

    if (tokensList.length === 0) {
      toast.error('请输入 Token');
      return;
    }

    setSubmitting(true);
    try {
      let result;
      switch (activeTab) {
        case 'audio':
          result = await api.createAudioTokens(tokensList);
          break;
        case 'ocr':
          result = await api.createOCRTokens(tokensList);
          break;
        case 'chat':
          result = await api.createChatTokens(tokensList);
          break;
      }
      toast.success(`成功导入 ${result!.created} 个 Token`);
      if (result!.duplicates > 0) {
        toast.warning(`${result!.duplicates} 个重复 Token 已跳过`);
      }
      setDialogOpen(false);
      setNewTokens('');
      fetchTokens();
    } catch (err) {
      toast.error('导入失败', {
        description: err instanceof Error ? err.message : '',
      });
    } finally {
      setSubmitting(false);
    }
  };

  // 删除单个
  const handleDelete = async (id: number) => {
    try {
      switch (activeTab) {
        case 'audio':
          await api.deleteAudioToken(id);
          break;
        case 'ocr':
          await api.deleteOCRToken(id);
          break;
        case 'chat':
          await api.deleteChatToken(id);
          break;
      }
      toast.success('删除成功');
      fetchTokens();
    } catch {
      toast.error('删除失败');
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedIds.size === 0) {
      toast.error('请选择要删除的 Token');
      return;
    }

    try {
      const ids = Array.from(selectedIds);
      switch (activeTab) {
        case 'audio':
          await api.batchDeleteAudioTokens(ids);
          break;
        case 'ocr':
          await api.batchDeleteOCRTokens(ids);
          break;
        case 'chat':
          await api.batchDeleteChatTokens(ids);
          break;
      }
      toast.success(`成功删除 ${ids.length} 个 Token`);
      fetchTokens();
    } catch {
      toast.error('批量删除失败');
    }
  };

  // 切换单个状态
  const handleToggle = async (id: number, currentEnabled: boolean) => {
    try {
      switch (activeTab) {
        case 'audio':
          await api.toggleAudioToken(id);
          break;
        case 'ocr':
          await api.toggleOCRToken(id);
          break;
        case 'chat':
          await api.toggleChatToken(id);
          break;
      }
      toast.success(currentEnabled ? '已禁用' : '已启用');
      fetchTokens();
    } catch {
      toast.error('操作失败');
    }
  };

  // 批量启用/禁用
  const handleBatchToggle = async (enable: boolean) => {
    if (selectedIds.size === 0) {
      toast.error('请选择要操作的 Token');
      return;
    }

    try {
      const ids = Array.from(selectedIds);
      switch (activeTab) {
        case 'audio':
          await api.batchToggleAudioTokens(ids, enable);
          break;
        case 'ocr':
          await api.batchToggleOCRTokens(ids, enable);
          break;
        case 'chat':
          await api.batchToggleChatTokens(ids, enable);
          break;
      }
      toast.success(`已${enable ? '启用' : '禁用'} ${ids.length} 个 Token`);
      fetchTokens();
    } catch {
      toast.error('批量操作失败');
    }
  };

  // 格式化日期
  const formatDate = (dateStr: string | null) => {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString('zh-CN');
  };

  const formatDateTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-neutral-950 p-6">
      <Toaster />

      <div className="mx-auto max-w-6xl space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Token 管理</h1>
          <Button variant="outline" onClick={handleLogout}>
            <LogOut className="mr-2 h-4 w-4" />
            退出登录
          </Button>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => router.push('/apikeys')}>
            <Key className="mr-1.5 h-4 w-4" />
            API Key
          </Button>
          <Button variant="outline" size="sm" onClick={() => router.push('/logs')}>
            <FileText className="mr-1.5 h-4 w-4" />
            请求日志
          </Button>
        </div>

        {/* Tabs */}
        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as TokenType)}>
          <TabsList>
            <TabsTrigger value="audio">Audio Token</TabsTrigger>
            <TabsTrigger value="ocr">OCR Token</TabsTrigger>
            <TabsTrigger value="chat">Chat Token</TabsTrigger>
          </TabsList>

          <TabsContent value={activeTab} className="space-y-4">
            {/* Toolbar */}
            <div className="flex flex-wrap items-center gap-2">
              <Button onClick={() => setDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                导入
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
            ) : tokens.length === 0 ? (
              <div className="text-center py-8 text-neutral-500">暂无数据</div>
            ) : (
              <div className="rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-12">
                        <Checkbox
                          checked={selectedIds.size === tokens.length && tokens.length > 0}
                          onCheckedChange={toggleSelectAll}
                        />
                      </TableHead>
                      <TableHead>ID</TableHead>
                      <TableHead>Token</TableHead>
                      <TableHead>导入时间</TableHead>
                      <TableHead>最后使用</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>总调用</TableHead>
                      <TableHead>今日调用</TableHead>
                      {activeTab === 'ocr' && <TableHead>每日限额</TableHead>}
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tokens.map((token) => (
                      <TableRow key={token.id}>
                        <TableCell>
                          <Checkbox
                            checked={selectedIds.has(token.id)}
                            onCheckedChange={() => toggleSelect(token.id)}
                          />
                        </TableCell>
                        <TableCell>{token.id}</TableCell>
                        <TableCell className="max-w-[200px] truncate font-mono text-sm">
                          {token.token}
                        </TableCell>
                        <TableCell>{formatDateTime(token.imported_at)}</TableCell>
                        <TableCell>{formatDate(token.last_used_at)}</TableCell>
                        <TableCell>
                          <span
                            className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
                              token.enabled
                                ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                            }`}
                          >
                            {token.enabled ? '启用' : '禁用'}
                          </span>
                        </TableCell>
                        <TableCell>{token.total_call_count}</TableCell>
                        <TableCell>{token.daily_call_count}</TableCell>
                        {activeTab === 'ocr' && (
                          <TableCell>
                            <span className="text-sm">
                              {token.daily_limit === 0 ? (
                                <span className="text-neutral-400">无限制</span>
                              ) : (
                                `${token.daily_call_count}/${token.daily_limit}`
                              )}
                            </span>
                          </TableCell>
                        )}
                        <TableCell>
                          <div className="flex items-center gap-1">
                            <Button
                              variant="ghost"
                              size="icon-xs"
                              onClick={() => handleToggle(token.id, token.enabled)}
                              title={token.enabled ? '禁用' : '启用'}
                            >
                              {token.enabled ? (
                                <PowerOff className="h-3.5 w-3.5" />
                              ) : (
                                <Power className="h-3.5 w-3.5" />
                              )}
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon-xs"
                              onClick={() => handleDelete(token.id)}
                              title="删除"
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
          </TabsContent>
        </Tabs>
      </div>

      {/* Import Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>导入 {activeTab.toUpperCase()} Token</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <Textarea
              placeholder="每行一个 Token"
              value={newTokens}
              onChange={(e) => setNewTokens(e.target.value)}
              rows={6}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleImport} disabled={submitting}>
              {submitting ? '导入中...' : '导入'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
