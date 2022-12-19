import ErrorModal from "../Modal/ErrorModal";
import { Component, ReactNode } from "react";

type ErrorBoundaryProps = {
  children: ReactNode;
};

class ErrorBoundary extends Component<ErrorBoundaryProps> {
  state = {
    error: null,
    errorInfo: null,
    modalOpen: false,
  };

  componentDidCatch(error, errorInfo) {
    // Catch errors in any components below and re-render with error message
    this.setState({
      error: error,
      errorInfo: errorInfo,
      modalOpen: true,
    });
  }

  render() {
    if (this.state.error && this.state.modalOpen) {
      return (
        <ErrorModal
          error={this.state.error}
          title="Sorry, but something went wrong"
        />
      );
    }
    return this.props.children;
  }
}

export default ErrorBoundary;
