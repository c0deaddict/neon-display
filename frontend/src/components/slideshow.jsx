import * as React from "react";
import ProgressBar from "./progress_bar.jsx";

export default class Slideshow extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      index: 0,
    };
  }

  getBackgroundColor() {
    return this.props.slideshow.background_color || "black";
  }

  getProgress() {
    return (this.state.index + 1) / this.props.slideshow.photos.length;
  }

  componentDidMount() {
    this.timer = setInterval(
      this.nextPhoto.bind(this),
      this.props.slideshow.delay_seconds * 1000
    );
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  nextPhoto() {
    const index = (this.state.index + 1) % this.props.slideshow.photos.length;
    this.setState({ index });
  }

  render() {
    const photo = this.props.slideshow.photos[this.state.index];
    return (
      <div
        className="slideshow"
        style={{ backgroundColor: this.getBackgroundColor() }}
      >
        <img src={"/photo/" + photo.image_path} />
        <ProgressBar progress={this.getProgress()} />
      </div>
    );
  }
}
