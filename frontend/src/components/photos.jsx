import * as React from "react";
import ProgressBar from "./progress_bar.jsx";

export default class Photos extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      index: 0,
    };
  }

  getBackgroundColor() {
    return this.props.data.background_color || "black";
  }

  getProgress() {
    return (this.state.index + 1) / this.props.data.photos.length;
  }

  componentDidMount() {
    this.timer = setInterval(
      this.nextPhoto.bind(this),
      this.props.data.delay_seconds * 1000
    );
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  nextPhoto() {
    const index = (this.state.index + 1) % this.props.data.photos.length;
    this.setState({ index });
  }

  pause() {
    console.error("not implemented");
  }

  resume() {
    console.error("not implemented");
  }

  render() {
    const photo = this.props.data.photos[this.state.index];
    return (
      <div
        className="photos"
        style={{ backgroundColor: this.getBackgroundColor() }}
      >
        <img src={"/photo/" + photo.image_path} />
        <ProgressBar progress={this.getProgress()} />
      </div>
    );
  }
}
