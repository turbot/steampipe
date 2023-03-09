import dayjs from "dayjs";

const timestampForFilename = (date: Date | number) => {
  const nowParsed = dayjs(date);
  return nowParsed.format("YYYYMMDDTHHmmss");
};

export { timestampForFilename };
