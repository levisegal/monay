"use client";

import { useEffect } from "react";

interface SlideOverProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  "data-testid"?: string;
}

export function SlideOver({ isOpen, onClose, title, children, "data-testid": testId }: SlideOverProps) {
  // Close on escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    if (isOpen) {
      document.addEventListener("keydown", handleEscape);
      document.body.style.overflow = "hidden";
    }
    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.body.style.overflow = "";
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-ink/30 transition-opacity"
        onClick={onClose}
      />

      {/* Panel */}
      <div data-testid={testId} className="fixed inset-y-0 right-0 w-full md:w-[576px] bg-white border-l border-paper-gray shadow-paper-lg overflow-y-auto">
        {/* Header */}
        <div className="sticky top-0 bg-white border-b border-paper-gray px-6 py-4 flex items-center justify-between">
          {title && (
            <h2 className="font-serif text-xl font-semibold text-ink">
              {title}
            </h2>
          )}
          <button
            onClick={onClose}
            className="p-2 hover:bg-paper-gray rounded-md transition-colors ml-auto"
            aria-label="Close"
          >
            <svg
              className="w-5 h-5 text-foreground-secondary"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {/* Content */}
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
}
