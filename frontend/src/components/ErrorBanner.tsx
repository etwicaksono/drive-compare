interface ErrorBannerProps {
  message: string;
  onDismiss?: () => void;
}

function ErrorBanner({ message, onDismiss }: ErrorBannerProps): React.JSX.Element {
  return (
    <div className="flex items-center justify-between rounded-lg border border-red-800/50 bg-red-900/20 px-4 py-3">
      <p className="text-sm text-red-400">{message}</p>
      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          className="ml-4 text-sm text-red-500 transition-colors hover:text-red-300"
        >
          Dismiss
        </button>
      )}
    </div>
  );
}

export default ErrorBanner;
