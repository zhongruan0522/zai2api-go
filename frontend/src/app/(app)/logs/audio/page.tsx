'use client';

import { api } from '@/lib/api';
import ChannelLogPage from '@/components/ChannelLogPage';

export default function AudioLogsPage() {
  return (
    <ChannelLogPage
      channel="audio"
      fetchLogs={(page, pageSize) => api.getAudioLogs(page, pageSize)}
      fetchStats={() => api.getAudioLogStats()}
    />
  );
}
