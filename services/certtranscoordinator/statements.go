package certtranscoordinator

var queryMap = map[string]string{
	"getServers": "select id, url, index, index_updated, step, tree_size, tree_size_updated from am.certificate_servers",
	"insertServer": `insert into am.certificate_servers (url, index, index_updated, step, tree_size, tree_size_updated) values 
	($1, $2, $3, $4, $5, $6) ON CONFLICT (url) DO UPDATE SET
		index=EXCLUDED.index,
		index_updated=EXCLUDED.index_updated,
		step=EXCLUDED.step,
		tree_size=EXCLUDED.tree_size,
		tree_size_updated=EXCLUDED.tree_size_updated
	 returning id`,
}
