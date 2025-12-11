import { useState } from 'react';

interface DeleteAccountModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void>;
  username: string;
}

const DeleteAccountModal = ({ isOpen, onClose, onConfirm, username }: DeleteAccountModalProps) => {
  const [confirmText, setConfirmText] = useState('');
  const [deleting, setDeleting] = useState(false);

  if (!isOpen) return null;

  const handleConfirm = async () => {
    if (confirmText !== username) return;
    
    try {
      setDeleting(true);
      await onConfirm();
    } finally {
      setDeleting(false);
    }
  };

  const handleClose = () => {
    if (!deleting) {
      setConfirmText('');
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white rounded-lg shadow-xl max-w-md w-full p-6">
          {/* Icon */}
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
            <svg
              className="h-6 w-6 text-red-600"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth="1.5"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
              />
            </svg>
          </div>

          {/* Content */}
          <div className="mt-3 text-center">
            <h3 className="text-lg font-semibold text-gray-900">
              Xóa tài khoản
            </h3>
            <div className="mt-2 text-sm text-gray-600 space-y-2">
              <p>
                Bạn có chắc chắn muốn xóa tài khoản của mình không?
              </p>
              <p className="font-semibold text-red-600">
                Hành động này không thể hoàn tác!
              </p>
              <p>
                Tất cả dữ liệu của bạn bao gồm farms, MCUs, survey points và lịch sử sensor sẽ bị xóa vĩnh viễn.
              </p>
            </div>

            {/* Confirmation Input */}
            <div className="mt-4">
              <label className="block text-sm text-left font-medium text-gray-700 mb-2">
                Để xác nhận, hãy nhập tên đăng nhập của bạn:{' '}
                <span className="font-semibold text-gray-900">{username}</span>
              </label>
              <input
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder={username}
                disabled={deleting}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent disabled:opacity-50 disabled:cursor-not-allowed"
              />
            </div>
          </div>

          {/* Buttons */}
          <div className="mt-6 flex gap-3">
            <button
              type="button"
              onClick={handleClose}
              disabled={deleting}
              className="flex-1 px-4 py-2 bg-gray-100 text-gray-900 rounded-lg hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Hủy
            </button>
            <button
              type="button"
              onClick={handleConfirm}
              disabled={confirmText !== username || deleting}
              className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {deleting && (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              )}
              {deleting ? 'Đang xóa...' : 'Xóa tài khoản'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DeleteAccountModal;
