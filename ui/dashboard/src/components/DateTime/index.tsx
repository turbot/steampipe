import dayjs from "dayjs";
import { classNames } from "../../utils/styles";

interface DateTimeProps {
  className?: string;
  date: dayjs.Dayjs | Date | string | number;
  dateClassName?: string;
  dateFormat?: string;
  timeClassName?: string;
  timeFormat?: string;
}

const DateTime = ({
  className,
  date,
  dateClassName,
  dateFormat = "D-MMM-YYYY",
  timeClassName,
  timeFormat = "HH:mm:ss",
}: DateTimeProps) => {
  const d = dayjs(date);
  return (
    <div className={classNames("tabular-nums space-x-1", className)}>
      <span className={classNames("text-foreground-lighter", dateClassName)}>
        {d.format(dateFormat)}
      </span>
      <span className={timeClassName}>{d.format(timeFormat)}</span>
    </div>
  );
};

export default DateTime;
