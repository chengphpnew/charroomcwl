/*
Navicat MySQL Data Transfer

Source Server         : 本地环境
Source Server Version : 50553
Source Host           : localhost:3306
Source Database       : thinkcmf5

Target Server Type    : MYSQL
Target Server Version : 50553
File Encoding         : 65001

Date: 2019-09-05 15:54:46
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for `im_hisstory`
-- ----------------------------
DROP TABLE IF EXISTS `im_hisstory`;
CREATE TABLE `im_hisstory` (
  `hisstory_id` int(11) NOT NULL AUTO_INCREMENT,
  `hisstory_content` varchar(300) NOT NULL COMMENT '群聊内容',
  `hisstory_room_name` varchar(50) NOT NULL COMMENT '群聊房间名称',
  `hisstory_time` int(10) NOT NULL COMMENT '群聊信息时间',
  `hisstory_username` varchar(20) NOT NULL COMMENT '群聊发信息人用户名',
  `hisstory_img` longblob,
  PRIMARY KEY (`hisstory_id`)
) ENGINE=InnoDB AUTO_INCREMENT=144 DEFAULT CHARSET=utf8 COMMENT='群聊信息记录表';

-- ----------------------------
-- Records of im_hisstory
-- ----------------------------
INSERT INTO `im_hisstory` VALUES ('1', '你好啊', '烟花爆竹', '1536996696', '成文龙', '');
