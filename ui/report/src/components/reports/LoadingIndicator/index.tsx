import Icon from "../../Icon";
import { loadingIcon } from "../../../constants/icons";

interface LoadingIndicatorProps {
  className?: string;
}

const LoadingIndicator = ({ className }: LoadingIndicatorProps) => (
  <Icon
    className={className ? className : undefined}
    icon={loadingIcon}
    spin={true}
  />
);

export default LoadingIndicator;
