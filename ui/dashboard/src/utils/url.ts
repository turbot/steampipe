const isRelativeUrl = (url) => {
  return (
    new URL(document.baseURI).origin === new URL(url, document.baseURI).origin
  );
};

export { isRelativeUrl };
