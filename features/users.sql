CREATE TABLE `funnel_users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `tid` bigint(20) NOT NULL DEFAULT 0,
  `name` varchar(64) NOT NULL DEFAULT 'anonymous',
  `tgname` varchar(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `index_tid` (`tid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;
