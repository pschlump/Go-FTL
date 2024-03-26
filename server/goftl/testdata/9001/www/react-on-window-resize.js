// From: https://www.hawatel.com/blog/handle-window-resize-in-react
// See also: Component with a de-bounceer: https://github.com/cesarandreu/react-window-resize-listener

import React, { Component } from 'react';
import LineChart from 'chart-graphs';

export default class Chart extends Component {

  constructor() {
    super();
    this.state = {
      width:  800,
      height: 182
    }
  }

  /**
   * Calculate & Update state of new dimensions
   */
  updateDimensions() {
    if(window.innerWidth < 500) {
      this.setState({ width: 450, height: 102 });
    } else {
      let update_width  = window.innerWidth-100;
      let update_height = Math.round(update_width/4.4);
      this.setState({ width: update_width, height: update_height });
    }
  }

  /**
   * Add event listener
   */
  componentDidMount() {
    this.updateDimensions();
    window.addEventListener("resize", this.updateDimensions.bind(this));
  }

  /**
   * Remove event listener
   */
  componentWillUnmount() {
    window.removeEventListener("resize", this.updateDimensions.bind(this));
  }

  render() {
    return(
      <div id="lineChart"> 
         <LineChart width={this.state.width} height={this.state.height} /> 
      </div>
    );
  }
}

/*
From: http://stackoverflow.com/questions/19014250/reactjs-rerender-on-browser-resize
var WindowDimensions = React.createClass({
    render: function() {
        return <span>{this.state.width} x {this.state.height}</span>;
    },
    updateDimensions: function() {
        this.setState({width: $(window).width(), height: $(window).height()});
    },
    componentWillMount: function() {
        this.updateDimensions();
    },
    componentDidMount: function() {
        window.addEventListener("resize", this.updateDimensions);
    },
    componentWillUnmount: function() {
        window.removeEventListener("resize", this.updateDimensions);
    }
});
*/
