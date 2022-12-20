import dayjs from "dayjs";
import { classNames } from "../../utils/styles";

type DateTimeProps = {
  className?: string;
  date: dayjs.Dayjs | Date | string | number;
  dateClassName?: string;
  dateFormat?: string;
  timeClassName?: string;
};

const DateTime = ({
  className,
  date,
  dateClassName,
  dateFormat = "D-MMM-YYYY",
  timeClassName,
}: DateTimeProps) => {
  const d = dayjs(date);
  return (
    <div className={classNames("tabular-nums space-x-1", className)}>
      <span className={classNames("text-foreground-lighter", dateClassName)}>
        {d.format(dateFormat)}
      </span>
      <span className={timeClassName}>{d.format("HH:mm:ss")}</span>
    </div>
  );
};

export default DateTime;
