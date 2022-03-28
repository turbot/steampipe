import Button, { ButtonProps } from "../Button";
import { classNames } from "../../../utils/styles";

const NeutralButton = ({
  children,
  className = "",
  disabled = false,
  onClick,
  size = "md",
  title,
  type,
}: ButtonProps) => (
  <Button
    className={classNames(
      "bg-dashboard-panel border border-black-scale-2 text-light hover:bg-black-scale-2 hover:border-black-scale-2 disabled:bg-dashboard disabled:text-light",
      className
    )}
    disabled={disabled}
    onClick={onClick}
    size={size}
    title={title}
    type={type}
  >
    {children}
  </Button>
);

export default NeutralButton;
