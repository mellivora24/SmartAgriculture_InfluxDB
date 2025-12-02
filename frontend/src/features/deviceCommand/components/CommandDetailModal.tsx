import { Fragment } from 'react';
import type { CommandInfo } from '@/core/types';

interface CommandDetailModalProps {
  command: CommandInfo | null;
  isOpen: boolean;
  onClose: () => void;
}

const CommandDetailModal = ({ command, isOpen, onClose }: CommandDetailModalProps) => {
  if (!isOpen || !command) return null;

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'text-green-600 bg-green-50 border-green-200';
      case 'failed':
        return 'text-red-600 bg-red-50 border-red-200';
      case 'pending':
        return 'text-yellow-600 bg-yellow-50 border-yellow-200';
      case 'sent':
        return 'text-blue-600 bg-blue-50 border-blue-200';
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZoneName: 'short'
    });
  };

  const getExecutionTime = () => {
    if (!command.executed_at || !command.created_at) return null;
    
    const created = new Date(command.created_at).getTime();
    const executed = new Date(command.executed_at).getTime();
    const diff = executed - created;
    
    if (diff < 1000) return `${diff}ms`;
    return `${(diff / 1000).toFixed(3)}s`;
  };

  return (
    <Fragment>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black bg-opacity-50 z-40 transition-opacity"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden">
          {/* Header */}
          <div className="bg-gray-50 px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <h2 className="text-xl font-semibold text-gray-900">Command Details</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Content */}
          <div className="px-6 py-4 overflow-y-auto max-h-[calc(90vh-120px)]">
            {/* Status Badge */}
            <div className="mb-6">
              <span className={`inline-flex items-center px-4 py-2 rounded-full text-sm font-semibold border ${getStatusColor(command.status)}`}>
                {command.status.toUpperCase()}
              </span>
            </div>

            {/* Command Info Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Command ID */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Command ID
                </label>
                <div className="px-3 py-2 bg-gray-50 rounded border border-gray-200">
                  <p className="text-sm font-mono text-gray-900 break-all">
                    {command.command_id}
                  </p>
                </div>
              </div>

              {/* Survey Point ID */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Survey Point ID
                </label>
                <div className="px-3 py-2 bg-gray-50 rounded border border-gray-200">
                  <p className="text-sm font-mono text-gray-900 break-all">
                    {command.survey_point_id}
                  </p>
                </div>
              </div>

              {/* Survey Point Name */}
              {command.survey_point_name && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Survey Point Name
                  </label>
                  <div className="px-3 py-2 bg-gray-50 rounded border border-gray-200">
                    <p className="text-sm text-gray-900">{command.survey_point_name}</p>
                  </div>
                </div>
              )}

              {/* Device Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Device Name
                </label>
                <div className="px-3 py-2 bg-gray-50 rounded border border-gray-200">
                  <p className="text-sm text-gray-900">{command.device_name}</p>
                </div>
              </div>

              {/* Command */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Command
                </label>
                <div className="px-3 py-2 bg-blue-50 rounded border border-blue-200">
                  <p className="text-sm font-semibold text-blue-900 uppercase">
                    {command.command}
                  </p>
                </div>
              </div>

              {/* Status */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Status
                </label>
                <div className={`px-3 py-2 rounded border ${getStatusColor(command.status)}`}>
                  <p className="text-sm font-semibold uppercase">
                    {command.status}
                  </p>
                </div>
              </div>
            </div>

            {/* Timestamps */}
            <div className="mt-6 pt-6 border-t border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Timeline</h3>
              
              <div className="space-y-4">
                {/* Created At */}
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <div className="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center">
                      <span className="text-blue-600">📤</span>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-900">Command Created</p>
                    <p className="text-xs text-gray-500">{formatDate(command.created_at)}</p>
                  </div>
                </div>

                {/* Executed At */}
                {command.executed_at && (
                  <div className="flex items-start">
                    <div className="flex-shrink-0">
                      <div className={`h-8 w-8 rounded-full flex items-center justify-center ${
                        command.status === 'success' ? 'bg-green-100' : 'bg-red-100'
                      }`}>
                        <span>{command.status === 'success' ? '✓' : '✗'}</span>
                      </div>
                    </div>
                    <div className="ml-4">
                      <p className="text-sm font-medium text-gray-900">Command Executed</p>
                      <p className="text-xs text-gray-500">{formatDate(command.executed_at)}</p>
                    </div>
                  </div>
                )}
              </div>

              {/* Execution Time */}
              {getExecutionTime() && (
                <div className="mt-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-blue-900">Execution Time</span>
                    <span className="text-sm font-semibold text-blue-600">
                      {getExecutionTime()}
                    </span>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="bg-gray-50 px-6 py-4 border-t border-gray-200 flex justify-end">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </Fragment>
  );
};

export default CommandDetailModal;
