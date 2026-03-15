'use client';

import { api } from '@/lib/api';
import ChannelLogPage from '@/components/ChannelLogPage';

export default function OCRLogsPage() {
  return (
    <ChannelLogPage
      channel="ocr"
      fetchLogs={(page, pageSize) => api.getOCRLogs(page, pageSize)}
      fetchStats={() => api.getOCRLogStats()}
    />
  );
}
