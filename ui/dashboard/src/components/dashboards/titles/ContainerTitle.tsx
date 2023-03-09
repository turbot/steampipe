import { classNames } from "../../../utils/styles";

const ContainerTitle = ({ title }) => {
  if (!title) {
    return null;
  }
  return <h2 className={classNames("col-span-12")}>{title}</h2>;
};

export default ContainerTitle;
