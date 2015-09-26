var data = {
	"copiers": [
		{
			"name": "my personal server",
			"type": "scp",
			"conf": {
				"user": "sphax",
				"host": "192.168.1.34",
				"port": 22
			}
		},
		{
			"name": "my google drive",
			"type": "gdrive",
			"conf": {
				"user": "example@gmail.com"
			}
		}
	]
};

React.render(
	<Container data={data} />,
	document.getElementById("container")
);