'use client';

import { useState, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { toast } from 'sonner';
import { api, TokenItem } from '@/lib/api';
import { Plus, Trash2, Power, PowerOff } from 'lucide-react';

type TokenType = 'audio' | 'ocr' | 'chat';

export default function TokensPage() {
  const [activeTab, setActiveTab] = useState<TokenType>('audio');
  const [tokens, setTokens] = useState<TokenItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [dialogOpen, setDialogOpen] = useState(false);
  const [newTokens, setNewTokens] = useState('');
  const [submitting, setSubmitting] = useState(false);

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
    fetchTokens();
  }, [activeTab, fetchTokens]);

  const toggleSelect = (id: number) => {
    const newSet = new Set(selectedIds);
    if (newSet.has(id)) newSet.delete(id);
    else newSet.add(id);
    setSelectedIds(newSet);
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === tokens.length) setSelectedIds(new Set());
    else setSelectedIds(new Set(tokens.map((t) => t.id)));
  };

  const BATCH_SIZE = 500;

  const createTokensBatch = async (channel: TokenType, tokensList: string[]) => {
    switch (channel) {
      case 'audio':
        return api.createAudioTokens(tokensList);
      case 'ocr':
        return api.createOCRTokens(tokensList);
      case 'chat':
        return api.createChatTokens(tokensList);
    }
  };

  const handleImport = async () => {
    const tokensList = newTokens.split('\n').map((t) => t.trim()).filter((t) => t);
    if (tokensList.length === 0) {
      toast.error('请输入 Token');
      return;
    }
    setSubmitting(true);

    try {
      const totalCreated = { value: 0 };
      const totalDuplicates = { value: 0 };
      const batches: string[][] = [];

      for (let i = 0; i < tokensList.length; i += BATCH_SIZE) {
        batches.push(tokensList.slice(i, i + BATCH_SIZE));
      }

      for (let i = 0; i < batches.length; i++) {
        const result = await createTokensBatch(activeTab, batches[i]);
        totalCreated.value += result!.created;
        totalDuplicates.value += result!.duplicates;
        if (batches.length > 1) {
          toast.info(`批次 ${i + 1}/${batches.length} 导入完成`);
        }
      }

      toast.success(`成功导入 ${totalCreated.value} 个 Token`);
      if (totalDuplicates.value > 0) {
        toast.warning(`${totalDuplicates.value} 个重复 Token 已跳过`);
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

  const formatDate = (dateStr: string | null) => {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString('zh-CN');
  };

  const formatDateTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  return (
    <div className="space-y-5">
      {/* Tabs for channel selection */}
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as TokenType)}>
        <TabsList>
          <TabsTrigger value="audio">Audio Token</TabsTrigger>
          <TabsTrigger value="ocr">OCR Token</TabsTrigger>
          <TabsTrigger value="chat">Chat Token</TabsTrigger>
        </TabsList>
      </Tabs>

      {/* Action bar */}
      <section className="rounded-3xl border border-border bg-card p-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-2">
            <Checkbox
              checked={selectedIds.size === tokens.length && tokens.length > 0}
              onCheckedChange={toggleSelectAll}
            />
            <span className="text-xs text-muted-foreground">
              已选 {selectedIds.size} / {tokens.length} 个
            </span>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <button
              className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-primary hover:text-primary disabled:cursor-not-allowed disabled:opacity-50"
              disabled={loading}
              onClick={fetchTokens}
            >
              刷新列表
            </button>
            <button
              className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-primary hover:text-primary"
              onClick={() => setDialogOpen(true)}
            >
              <Plus className="inline h-4 w-4 mr-1" />
              导入 Token
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

        {/* Token table */}
        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-muted-foreground">
                <th className="px-3 py-2 w-10"></th>
                <th className="px-3 py-2">ID</th>
                <th className="px-3 py-2">Token</th>
                <th className="px-3 py-2">导入时间</th>
                <th className="px-3 py-2">最后使用</th>
                <th className="px-3 py-2">状态</th>
                <th className="px-3 py-2">总调用</th>
                <th className="px-3 py-2">今日调用</th>
                {activeTab === 'ocr' && <th className="px-3 py-2">每日限额</th>}
                <th className="px-3 py-2">操作</th>
              </tr>
            </thead>
            <tbody>
              {tokens.map((token) => (
                <tr key={token.id} className="border-b border-border/50">
                  <td className="px-3 py-2">
                    <Checkbox
                      checked={selectedIds.has(token.id)}
                      onCheckedChange={() => toggleSelect(token.id)}
                    />
                  </td>
                  <td className="px-3 py-2 text-foreground">{token.id}</td>
                  <td className="px-3 py-2 font-mono text-xs max-w-[200px] truncate text-foreground">
                    {token.token}
                  </td>
                  <td className="px-3 py-2 text-muted-foreground whitespace-nowrap">{formatDateTime(token.imported_at)}</td>
                  <td className="px-3 py-2 text-muted-foreground">{formatDate(token.last_used_at)}</td>
                  <td className="px-3 py-2">
                    <span className={token.enabled ? 'text-emerald-600' : 'text-destructive'}>
                      {token.enabled ? '启用' : '禁用'}
                    </span>
                  </td>
                  <td className="px-3 py-2 text-foreground">{token.total_call_count}</td>
                  <td className="px-3 py-2 text-foreground">{token.daily_call_count}</td>
                  {activeTab === 'ocr' && (
                    <td className="px-3 py-2 text-foreground">
                      {token.daily_limit === 0 ? (
                        <span className="text-muted-foreground">无限制</span>
                      ) : (
                        `${token.daily_call_count}/${token.daily_limit}`
                      )}
                    </td>
                  )}
                  <td className="px-3 py-2">
                    <div className="flex items-center gap-1">
                      <button
                        className="rounded-xl border border-border px-3 py-1 text-xs text-foreground transition-colors hover:bg-accent"
                        onClick={() => handleToggle(token.id, token.enabled)}
                      >
                        {token.enabled ? '禁用' : '启用'}
                      </button>
                      <button
                        className="rounded-xl border border-destructive/30 px-3 py-1 text-xs text-destructive transition-colors hover:bg-destructive/10"
                        onClick={() => handleDelete(token.id)}
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
          {!loading && tokens.length === 0 && (
            <div className="py-8 text-center text-sm text-muted-foreground">暂无数据</div>
          )}
        </div>
      </section>

      {/* Import Dialog */}
      {dialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 px-4" onClick={() => !submitting && setDialogOpen(false)}>
          <div className="w-full max-w-lg rounded-3xl border border-border bg-card p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-foreground">导入 {activeTab.toUpperCase()} Token</p>
              <button className="text-xs text-muted-foreground hover:text-foreground" onClick={() => setDialogOpen(false)}>关闭</button>
            </div>
            <div className="mt-4">
              <Textarea
                placeholder="每行一个 Token"
                value={newTokens}
                onChange={(e) => setNewTokens(e.target.value)}
                rows={8}
                className="min-h-0 rounded-2xl border border-input bg-background px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <p className="mt-2 text-xs text-muted-foreground">
                {newTokens.split('\n').filter((t) => t.trim()).length > 0 && (
                  <>已输入 {newTokens.split('\n').filter((t) => t.trim()).length} 个 Token，超过 500 将自动分批导入</>
                )}
              </p>
            </div>
            <div className="mt-4 flex items-center justify-end gap-2">
              <button
                className="rounded-full border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-accent"
                onClick={() => setDialogOpen(false)}
                disabled={submitting}
              >
                取消
              </button>
              <button
                className="rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
                onClick={handleImport}
                disabled={submitting}
              >
                {submitting ? '导入中...' : '导入'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
