import { classNames } from "../../../utils/styles";

const DashboardTitle = ({ title }) => {
  if (!title) {
    return null;
  }
  return <h1 className={classNames("col-span-12")}>{title}</h1>;
};

export default DashboardTitle;
