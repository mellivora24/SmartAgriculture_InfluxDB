import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { commandApi } from '../api/commandApi';
import CommandDetailModal from '../components/CommandDetailModal';
import type { CommandInfo } from '@/core/types';

interface FilterState {
  device_name: string;
  status: string;
  search: string;
}

const CommandHistoryPage = () => {
  const { surveyPointId } = useParams<{ surveyPointId: string }>();
  const navigate = useNavigate();
  const [commands, setCommands] = useState<CommandInfo[]>([]);
  const [filteredCommands, setFilteredCommands] = useState<CommandInfo[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterState>({
    device_name: '',
    status: '',
    search: ''
  });
  const [selectedCommand, setSelectedCommand] = useState<CommandInfo | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Load command history
  useEffect(() => {
    const loadCommands = async () => {
      if (!surveyPointId) {
        setError('Survey Point ID is required');
        setIsLoading(false);
        return;
      }

      try {
        setIsLoading(true);
        setError(null);
        
        const response = await commandApi.getCommandHistory({
          survey_point_id: surveyPointId,
          limit: 100
        });

        if (response.data) {
          setCommands(response.data);
          setFilteredCommands(response.data);
        }
      } catch (err) {
        console.error('Failed to load command history:', err);
        setError('Failed to load command history');
      } finally {
        setIsLoading(false);
      }
    };

    loadCommands();
  }, [surveyPointId]);

  // Apply filters
  useEffect(() => {
    let filtered = [...commands];

    // Filter by device name
    if (filter.device_name) {
      filtered = filtered.filter(cmd => 
        cmd.device_name === filter.device_name
      );
    }

    // Filter by status
    if (filter.status) {
      filtered = filtered.filter(cmd => 
        cmd.status === filter.status
      );
    }

    // Search filter
    if (filter.search) {
      const searchLower = filter.search.toLowerCase();
      filtered = filtered.filter(cmd =>
        cmd.device_name.toLowerCase().includes(searchLower) ||
        cmd.command.toLowerCase().includes(searchLower) ||
        cmd.command_id.toLowerCase().includes(searchLower)
      );
    }

    setFilteredCommands(filtered);
  }, [filter, commands]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'failed':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'pending':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'sent':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return '✓';
      case 'failed':
        return '✗';
      case 'pending':
        return '⏳';
      case 'sent':
        return '📤';
      default:
        return '?';
    }
  };

  const getCommandIcon = (command: string) => {
    if (command === 'on' || command === 'turn_on') return '🔋';
    if (command === 'off' || command === 'turn_off') return '⏸️';
    return '🎛️';
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  const getExecutionTime = (cmd: CommandInfo) => {
    if (!cmd.executed_at || !cmd.created_at) return 'N/A';
    
    const created = new Date(cmd.created_at).getTime();
    const executed = new Date(cmd.executed_at).getTime();
    const diff = executed - created;
    
    if (diff < 1000) return `${diff}ms`;
    return `${(diff / 1000).toFixed(2)}s`;
  };

  const handleViewDetails = (cmd: CommandInfo) => {
    setSelectedCommand(cmd);
    setIsModalOpen(true);
  };

  // Get unique device names for filter
  const deviceNames = Array.from(new Set(commands.map(cmd => cmd.device_name)));

  // Statistics
  const stats = {
    total: commands.length,
    success: commands.filter(c => c.status === 'success').length,
    failed: commands.filter(c => c.status === 'failed').length,
    pending: commands.filter(c => c.status === 'pending').length
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Loading command history...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 text-lg mb-4">{error}</p>
          <button
            onClick={() => navigate(`/dashboard/${surveyPointId}`)}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <button
            onClick={() => navigate(`/dashboard/${surveyPointId}`)}
            className="mb-4 flex items-center text-sm text-gray-600 hover:text-gray-900"
          >
            <svg className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Dashboard
          </button>
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Command History</h1>
              <p className="mt-1 text-sm text-gray-500">
                Control command history for this survey point
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Statistics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Total Commands</p>
                <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
              </div>
              <div className="text-3xl">📊</div>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Successful</p>
                <p className="text-2xl font-bold text-green-600">{stats.success}</p>
              </div>
              <div className="text-3xl">✓</div>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Failed</p>
                <p className="text-2xl font-bold text-red-600">{stats.failed}</p>
              </div>
              <div className="text-3xl">✗</div>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Pending</p>
                <p className="text-2xl font-bold text-yellow-600">{stats.pending}</p>
              </div>
              <div className="text-3xl">⏳</div>
            </div>
          </div>
        </div>

        {/* Filters */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Filters</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Search */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Search
              </label>
              <input
                type="text"
                placeholder="Search by device, command, or ID..."
                value={filter.search}
                onChange={(e) => setFilter({ ...filter, search: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            {/* Device Name */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Device
              </label>
              <select
                value={filter.device_name}
                onChange={(e) => setFilter({ ...filter, device_name: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Devices</option>
                {deviceNames.map(name => (
                  <option key={name} value={name}>{name}</option>
                ))}
              </select>
            </div>

            {/* Status */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Status
              </label>
              <select
                value={filter.status}
                onChange={(e) => setFilter({ ...filter, status: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="">All Statuses</option>
                <option value="success">Success</option>
                <option value="failed">Failed</option>
                <option value="pending">Pending</option>
                <option value="sent">Sent</option>
              </select>
            </div>
          </div>

          {/* Clear Filters */}
          {(filter.search || filter.device_name || filter.status) && (
            <button
              onClick={() => setFilter({ device_name: '', status: '', search: '' })}
              className="mt-4 text-sm text-blue-600 hover:text-blue-700 font-medium"
            >
              Clear all filters
            </button>
          )}
        </div>

        {/* Command List */}
        <div className="bg-white rounded-lg shadow overflow-hidden">
          {filteredCommands.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-6xl mb-4">📭</div>
              <p className="text-gray-500 text-lg">No commands found</p>
              {(filter.search || filter.device_name || filter.status) && (
                <p className="text-sm text-gray-400 mt-2">Try adjusting your filters</p>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Time
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Device
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Command
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Execution Time
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {filteredCommands.map((cmd) => (
                    <tr key={cmd.command_id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900">
                          {formatDate(cmd.created_at)}
                        </div>
                        {cmd.executed_at && (
                          <div className="text-xs text-gray-500">
                            Executed: {formatDate(cmd.executed_at)}
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {cmd.device_name}
                        </div>
                        {cmd.survey_point_name && (
                          <div className="text-xs text-gray-500">
                            {cmd.survey_point_name}
                          </div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center">
                          <span className="text-xl mr-2">{getCommandIcon(cmd.command)}</span>
                          <span className="text-sm font-semibold text-gray-900 uppercase">
                            {cmd.command}
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold border ${getStatusColor(cmd.status)}`}>
                          <span className="mr-1">{getStatusIcon(cmd.status)}</span>
                          {cmd.status.toUpperCase()}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {getExecutionTime(cmd)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        <button
                          onClick={() => handleViewDetails(cmd)}
                          className="text-blue-600 hover:text-blue-800 font-medium"
                        >
                          View Details
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Results count */}
        {filteredCommands.length > 0 && (
          <div className="mt-4 text-center text-sm text-gray-500">
            Showing {filteredCommands.length} of {commands.length} commands
          </div>
        )}
      </div>

      {/* Command Detail Modal */}
      <CommandDetailModal
        command={selectedCommand}
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setSelectedCommand(null);
        }}
      />
    </div>
  );
};

export default CommandHistoryPage;
