'use client';

import { api } from '@/lib/api';
import ChannelLogPage from '@/components/ChannelLogPage';

export default function ChatLogsPage() {
  return (
    <ChannelLogPage
      channel="chat"
      fetchLogs={(page, pageSize) => api.getChatLogs(page, pageSize)}
      fetchStats={() => api.getChatLogStats()}
    />
  );
}
