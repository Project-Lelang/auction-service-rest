DELETE t1 FROM user_fcm_tokens t1
INNER JOIN user_fcm_tokens t2
    ON t1.fcm_token = t2.fcm_token
   AND t1.id < t2.id;

ALTER TABLE user_fcm_tokens
    ADD UNIQUE KEY uq_user_fcm_tokens_fcm_token (fcm_token);
