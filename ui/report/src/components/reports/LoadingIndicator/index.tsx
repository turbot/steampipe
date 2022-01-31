import Icon from "../../Icon";
import { loadingIcon } from "../../../constants/icons";
import { classNames } from "../../../utils/styles";

interface LoadingIndicatorProps {
  className?: string;
}

const LoadingIndicator = ({ className }: LoadingIndicatorProps) => (
  <Icon
    className={classNames(className, "text-black-scale-4")}
    icon={loadingIcon}
    spin={true}
  />
);

export default LoadingIndicator;
