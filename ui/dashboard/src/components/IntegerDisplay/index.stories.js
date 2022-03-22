import IntegerDisplay from "./index";

const story = {
  title: "Utilities/Integer Display",
  component: IntegerDisplay,
};

export default story;

const lessThan1k = {
  num: 796,
};
const moreThan1k = {
  num: 2134,
};
const moreThan10k = {
  num: 21340,
};
const moreThan100k = {
  num: 156467,
};
const moreThan1M = {
  num: 2000999,
};
const moreThan10M = {
  num: 21340000,
};
const moreThan1000M = {
  num: 2134000000,
};
const negativeNumber = {
  num: -123,
};
const numberString = {
  num: "9876",
};
const numberNull = {
  num: null,
};
const numberUndefined = {
  num: undefined,
};

export const LessThan1k = () => <IntegerDisplay {...lessThan1k} />;
export const MoreThan1k = () => <IntegerDisplay {...moreThan1k} />;
export const MoreThan10k = () => <IntegerDisplay {...moreThan10k} />;
export const MoreThan100k = () => <IntegerDisplay {...moreThan100k} />;
export const MoreThan1M = () => <IntegerDisplay {...moreThan1M} />;
export const MoreThan10M = () => <IntegerDisplay {...moreThan10M} />;
export const MoreThan1000M = () => <IntegerDisplay {...moreThan1000M} />;
export const NegativeNumber = () => <IntegerDisplay {...negativeNumber} />;
export const NumberAsString = () => <IntegerDisplay {...numberString} />;
export const NumberNull = () => <IntegerDisplay {...numberNull} />;
export const NumberUndefined = () => <IntegerDisplay {...numberUndefined} />;

export const LevelKAndLessThan1k = () => (
  <IntegerDisplay {...lessThan1k} startAt="k" />
);
export const LevelKAndMoreThan1k = () => (
  <IntegerDisplay {...moreThan1k} startAt="k" />
);
export const LevelKAndMoreThan10k = () => (
  <IntegerDisplay {...moreThan10k} startAt="k" />
);
export const LevelKAndMoreThan100k = () => (
  <IntegerDisplay {...moreThan100k} startAt="k" />
);
export const LevelKAndMoreThan1M = () => (
  <IntegerDisplay {...moreThan1M} startAt="k" />
);
export const LevelKAndMoreThan10M = () => (
  <IntegerDisplay {...moreThan10M} startAt="k" />
);
export const LevelKAndMoreThan1000M = () => (
  <IntegerDisplay {...moreThan1000M} startAt="k" />
);
export const LevelKAndNegativeNumber = () => (
  <IntegerDisplay {...negativeNumber} startAt="k" />
);
export const LevelKAndNumberAsString = () => (
  <IntegerDisplay {...numberString} startAt="k" />
);
export const LevelKAndNumberNull = () => (
  <IntegerDisplay {...numberNull} startAt="k" />
);
export const LevelKAndNumberUndefined = () => (
  <IntegerDisplay {...numberUndefined} startAt="k" />
);

export const LevelMAndLessThan1k = () => (
  <IntegerDisplay {...lessThan1k} startAt="m" />
);
export const LevelMAndMoreThan1k = () => (
  <IntegerDisplay {...moreThan1k} startAt="m" />
);
export const LevelMAndMoreThan10k = () => (
  <IntegerDisplay {...moreThan10k} startAt="m" />
);
export const LevelMAndMoreThan100k = () => (
  <IntegerDisplay {...moreThan100k} startAt="m" />
);
export const LevelMAndMoreThan1M = () => (
  <IntegerDisplay {...moreThan1M} startAt="m" />
);
export const LevelMAndMoreThan10M = () => (
  <IntegerDisplay {...moreThan10M} startAt="m" />
);
export const LevelMAndMoreThan1000M = () => (
  <IntegerDisplay {...moreThan1000M} startAt="m" />
);
export const LevelMAndNegativeNumber = () => (
  <IntegerDisplay {...negativeNumber} startAt="m" />
);
export const LevelMAndNumberAsString = () => (
  <IntegerDisplay {...numberString} startAt="m" />
);
export const LevelMAndNumberNull = () => (
  <IntegerDisplay {...numberNull} startAt="m" />
);
export const LevelMAndNumberUndefined = () => (
  <IntegerDisplay {...numberUndefined} startAt="m" />
);

export const LevelInfinityAndLessThan1k = () => (
  <IntegerDisplay {...lessThan1k} startAt={false} />
);
export const LevelInfinityAndMoreThan1k = () => (
  <IntegerDisplay {...moreThan1k} startAt={false} />
);
export const LevelInfinityAndMoreThan10k = () => (
  <IntegerDisplay {...moreThan10k} startAt={false} />
);
export const LevelInfinityAndMoreThan100k = () => (
  <IntegerDisplay {...moreThan100k} startAt={false} />
);
export const LevelInfinityAndMoreThan1M = () => (
  <IntegerDisplay {...moreThan1M} startAt={false} />
);
export const LevelInfinityAndMoreThan10M = () => (
  <IntegerDisplay {...moreThan10M} startAt={false} />
);
export const LevelInfinityAndMoreThan1000M = () => (
  <IntegerDisplay {...moreThan1000M} startAt={false} />
);
export const LevelInfinityAndNegativeNumber = () => (
  <IntegerDisplay {...negativeNumber} startAt={false} />
);
export const LevelInfinityAndNumberAsString = () => (
  <IntegerDisplay {...numberString} startAt={false} />
);
export const LevelInfinityAndNumberNull = () => (
  <IntegerDisplay {...numberNull} startAt={false} />
);
export const LevelInfinityAndNumberUndefined = () => (
  <IntegerDisplay {...numberUndefined} startAt={false} />
);
