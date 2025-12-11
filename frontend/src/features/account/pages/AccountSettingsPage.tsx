import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/core/hooks/useAuth';
import { userApi } from '../api/userApi';
import type { User, UpdateUserRequest } from '@/core/types/user';
import DeleteAccountModal from '../components/DeleteAccountModal';

const AccountSettingsPage = () => {
  const navigate = useNavigate();
  const { logout } = useAuth();
  
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  const [formData, setFormData] = useState({
    email: '',
    phone: '',
    full_name: '',
  });

  useEffect(() => {
    loadUserData();
  }, []);

  const loadUserData = async () => {
    try {
      setLoading(true);
      const response = await userApi.getCurrentUser();
      const userData = response.data;
      setUser(userData ?? null);
      setFormData({
        email: userData?.email || '',
        phone: userData?.phone || '',
        full_name: userData?.full_name || '',
      });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      console.error('Error loading user data:', err);
      setError(err.response?.data?.error || err.message || 'Unable to load user information');
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    setError('');
    setSuccessMessage('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;

    if (!formData.email.trim()) {
      setError('Email cannot be empty');
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(formData.email)) {
      setError('Invalid email');
      return;
    }

    try {
      setSaving(true);
      setError('');
      setSuccessMessage('');

      const updateData: UpdateUserRequest = {
        email: formData.email,
        phone: formData.phone || undefined,
        full_name: formData.full_name || undefined,
      };

      const response = await userApi.updateUser(user.id, updateData);
      const updatedUser = response.data;
      setUser(updatedUser ?? null);
      setSuccessMessage('Profile updated successfully!');
      setTimeout(() => setSuccessMessage(''), 3000);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      console.error('Error updating user:', err);
      setError(err.response?.data?.error || err.message || 'Error updating profile');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteAccount = async () => {
    if (!user) return;
    try {
      await userApi.deleteUser(user.id);
      logout();
      navigate('/login', { replace: true });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      console.error('Error deleting account:', err);
      setError(err.response?.data?.error || err.message || 'Error deleting account');
      setShowDeleteModal(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Unable to load user information</p>
          <button
            onClick={() => navigate('/')}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Back to Home
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-gray-900">Account Settings</h1>
          <p className="mt-1 text-sm text-gray-500">Manage your account information</p>
        </div>

        {successMessage && (
          <div className="mb-6 bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg">
            {successMessage}
          </div>
        )}

        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg">
            {error}
          </div>
        )}

        <div className="bg-white rounded-lg shadow mb-6">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">Account Information</h2>
          </div>
          <div className="px-6 py-4">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
                <input
                  type="text"
                  value={user.username}
                  disabled
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500 cursor-not-allowed"
                />
                <p className="mt-1 text-xs text-gray-500">Username cannot be changed</p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Account Created</label>
                <input
                  type="text"
                  value={new Date(user.created_at).toLocaleDateString('en-US', {
                    year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit'
                  })}
                  disabled
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500 cursor-not-allowed"
                />
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow mb-6">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">Edit Profile</h2>
          </div>
          <form onSubmit={handleSubmit} className="px-6 py-4">
            <div className="space-y-4">
              <div>
                <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
                  Email <span className="text-red-500">*</span>
                </label>
                <input
                  type="email"
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleInputChange}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
              </div>

              <div>
                <label htmlFor="full_name" className="block text-sm font-medium text-gray-700 mb-1">Full Name</label>
                <input
                  type="text"
                  id="full_name"
                  name="full_name"
                  value={formData.full_name}
                  onChange={handleInputChange}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              <div>
                <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-1">Phone Number</label>
                <input
                  type="tel"
                  id="phone"
                  name="phone"
                  value={formData.phone}
                  onChange={handleInputChange}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="0123456789"
                />
              </div>
            </div>

            <div className="mt-6 flex justify-end">
              <button
                type="submit"
                disabled={saving}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
              >
                {saving && <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>}
                {saving ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </form>
        </div>

        <div className="bg-white rounded-lg shadow border-2 border-red-200">
          <div className="px-6 py-4 bg-red-50 border-b border-red-200">
            <h2 className="text-lg font-semibold text-red-900">Danger Zone</h2>
          </div>
          <div className="px-6 py-4">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-sm font-semibold text-gray-900 mb-1">Delete Account</h3>
                <p className="text-sm text-gray-600">Once deleted, your account and all related data will be permanently removed and cannot be recovered.</p>
              </div>
              <button
                onClick={() => setShowDeleteModal(true)}
                className="ml-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 whitespace-nowrap"
              >
                Delete Account
              </button>
            </div>
          </div>
        </div>
      </div>

      <DeleteAccountModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={handleDeleteAccount}
        username={user.username}
      />
    </div>
  );
};

export default AccountSettingsPage;
