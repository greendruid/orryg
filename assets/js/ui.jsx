class SCPCopier {
	render() {
		return (
			<div className="row copier-row">
				<div className="three columns">
					<h5>{this.props.data.name}</h5>
				</div>
				<div className="two columns">{this.props.data.type}</div>
				<div className="three columns">{this.props.data.conf.user}</div>
				<div className="two columns">{this.props.data.conf.host}</div>
				<div className="two columns">{this.props.data.conf.port}</div>				
			</div>			
		);
	}
}

class GDriveCopier {
	render() {
		return (
			<div className="row copier-row">
				<div className="three columns">
					<h5>{this.props.data.name}</h5>
				</div>
				<div className="two columns">{this.props.data.type}</div>
				<div className="seven columns">{this.props.data.conf.user}</div>
			</div>		
		);
	}
}

class Header {
	render() {
		return (
			<div className="row header">
				<h1>Orryg</h1>
			</div>
		);
	}
}

class CopierRow {
	render() {
		if (this.props.data.type === "scp") {
			return <SCPCopier data={this.props.data} />;
		} else if (this.props.data.type === "gdrive") {
			return <GDriveCopier data={this.props.data} />;
		}
		return <div></div>;
	}
}

class CopierList {
	render() {
		var nodes = this.props.data.map(el => {
			return (
				<CopierRow key={el.name} data={el} />
			);
		});
		return (
			<div className="row copier-list">
				<div className="row">
					<h2 class="twelve columns">Copiers</h2>
				</div>
				<div className="row">
					{nodes}
				</div>
			</div>
		);
	}
}

class Container {
	render() {
		return (
			<div className="container">
				<Header />
				<CopierList data={this.props.data.copiers} />
			</div>
		); 
	}
}