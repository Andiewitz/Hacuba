interface ParubaLogoProps {
  iconOnly?: boolean;
  className?: string;
}

export default function ParubaLogo({
  iconOnly = false,
  className = "",
}: ParubaLogoProps) {
  return (
    <span className={`inline-flex items-center gap-1 ${className}`}>
      <svg
        width="32"
        height="32"
        viewBox="0 0 32 32"
        fill="none"
        className="shrink-0"
      >
        <path
          d="M16 2C8.268 2 2 8.268 2 16s6.268 14 14 14 14-6.268 14-14S23.732 2 16 2zm0 2c2.8 0 5.36 1.04 7.32 2.76L16 16V4zm-9.32 4.76A11.94 11.94 0 0116 4v12L6.68 6.76zM16 28a11.94 11.94 0 01-9.32-4.24L16 16v12zm2-12l9.32-7.24A11.94 11.94 0 0116 28V16z"
          fill="currentColor"
        />
      </svg>
      {!iconOnly && (
        <span className="text-xl font-semibold tracking-tight text-foreground font-heading">
          Paruba
        </span>
      )}
    </span>
  );
}
