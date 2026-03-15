'use client';

import { api } from '@/lib/api';
import ChannelLogPage from '@/components/ChannelLogPage';

export default function ImageLogsPage() {
  return (
    <ChannelLogPage
      channel="image"
      fetchLogs={(page, pageSize) => api.getImageLogs(page, pageSize)}
      fetchStats={() => api.getImageLogStats()}
    />
  );
}
