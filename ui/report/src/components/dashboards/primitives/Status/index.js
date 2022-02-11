import Icon from "../../../Icon";
import Primitive from "../../Primitive";
import PropTypes from "prop-types";
import {
  alertIcon,
  emptyIcon,
  infoIcon,
  okIcon,
} from "../../../../constants/icons";

const getIcon = (type) => {
  switch (type) {
    case "alarm":
      return alertIcon;
    case "info":
      return infoIcon;
    case "ok":
      return okIcon;
    default:
      return emptyIcon;
  }
};

const getIconClasses = (type) => {
  switch (type) {
    case "alarm":
      return "text-alert";
    case "ok":
      return "text-ok";
    default:
      return "d-none";
  }
};

const Status = ({ data, error, title }) => {
  if (!data || error) {
    return <Primitive error={error} ready={false} />;
  }
  const status = data[0][0];
  return (
    <Primitive error={error} ready={!!data}>
      <div className="text-6xl py-2 px-3">
        <Icon className={getIconClasses(status)} icon={getIcon(status)} />
      </div>
    </Primitive>
  );
};

Status.propTypes = {
  data: PropTypes.arrayOf(PropTypes.array),
  error: PropTypes.string,
  loading: PropTypes.bool,
  type: PropTypes.oneOf(["alert", "default", "info", "ok"]),
};

export default {
  type: "status",
  component: Status,
};
